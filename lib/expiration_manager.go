package redisgklib

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// listenerKeyEventManager - manager for working with key expiration notifications
type listenerKeyEventManager struct {
	client       *redis.Client
	ctx          context.Context
	cancel       context.CancelFunc
	keyEventChan chan KeyEvent
	mu           sync.RWMutex
	isRunning    bool
	wg           sync.WaitGroup // Add WaitGroup for proper goroutine completion
}

// newListenerKeyEventManager creates a new key expiration notification manager
func newListenerKeyEventManager(client *redis.Client, ctx context.Context) *listenerKeyEventManager {
	if client == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	managerCtx, cancel := context.WithCancel(ctx)

	return &listenerKeyEventManager{
		client:       client,
		ctx:          managerCtx,
		cancel:       cancel,
		keyEventChan: make(chan KeyEvent), // Unbuffered channel for simple forwarding
		isRunning:    false,
	}
}

// start starts the key  notification listener
func (em *listenerKeyEventManager) start() error {
	if em == nil {
		return fmt.Errorf("listener key event manager is nil")
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	if em.isRunning {
		// If listener is already running, just return success
		return nil
	}

	// Subscribe to all Redis event channels for proper event type detection
	channels := []string{
		"__keyevent@0__:expire", // Expiration events
		"__keyevent@0__:set",    // Creation/update events
		"__keyevent@0__:del",    // Deletion events
	}

	// Create subscription to key event notification channels
	pubsub := em.client.Subscribe(em.ctx, channels...)

	// Start goroutine for processing notifications
	em.wg.Add(1)
	go em.listenForEvents(pubsub)

	em.isRunning = true
	return nil
}

// listenForEvents listens for key event notifications
func (em *listenerKeyEventManager) listenForEvents(pubsub *redis.PubSub) {
	defer func() {
		pubsub.Close()
		em.wg.Done()
	}()

	for {
		select {
		case <-em.ctx.Done():
			return
		case msg := <-pubsub.Channel():
			event := em.processEventMessage(msg)
			if event.EventType != EventTypeUnknown {
				// Simply forward event to user (block until user reads)
				select {
				case em.keyEventChan <- event:
				case <-em.ctx.Done():
					return
				}
			}
		}
	}
}

// processEventMessage processes event message and determines event type by channel
func (em *listenerKeyEventManager) processEventMessage(msg *redis.Message) KeyEvent {
	var eventType EventType
	var key string

	switch msg.Channel {
	case "__keyevent@0__:expire":
		// Check if this is actually an expiration or just TTL setting
		if em.isKeyActuallyExpired(msg.Payload) {
			eventType = EventTypeExpired
		} else {
			// This is likely a TTL setting event, treat as creation/update
			eventType = EventTypeCreated
		}
		key = msg.Payload
	case "__keyevent@0__:set":
		eventType = EventTypeCreated // or Updated, depends on context
		key = msg.Payload
	case "__keyevent@0__:del":
		eventType = EventTypeDeleted
		key = msg.Payload
	default:
		eventType = EventTypeUnknown
		key = msg.Payload
	}

	// Get key value if possible
	value := ""
	if eventType == EventTypeExpired {
		// For expiring keys, try to get value before deletion
		value, _ = em.getKeyValue(key)
	} else if eventType == EventTypeCreated || eventType == EventTypeUpdated {
		// For created/updated keys, get current value
		value, _ = em.getKeyValue(key)
	}

	now := time.Now().UTC()

	return KeyEvent{
		Key:       key,
		Value:     value,
		EventType: eventType,
		Timestamp: now,
	}
}

// isKeyActuallyExpired checks if a key is actually expired or just has TTL set
func (em *listenerKeyEventManager) isKeyActuallyExpired(key string) bool {
	if em == nil || em.client == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(em.ctx, 100*time.Millisecond)
	defer cancel()

	// Check if key exists and get its TTL
	ttl, err := em.client.TTL(ctx, key).Result()
	if err != nil {
		// If we can't get TTL, assume it's not expired
		return false
	}

	// If TTL is -1, key has no expiration
	if ttl == -1 {
		return false
	}

	// If TTL is -2, key doesn't exist (already expired)
	if ttl == -2 {
		return true
	}

	// If TTL is positive, key still exists and hasn't expired yet
	if ttl > 0 {
		return false
	}

	// If TTL is 0, key should be expired
	if ttl == 0 {
		return true
	}

	return false
}

// stop stops the notification listener
func (em *listenerKeyEventManager) stop() {
	if em == nil {
		return
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	if !em.isRunning {
		return
	}

	// Cancel context
	if em.cancel != nil {
		em.cancel()
	}

	// Wait for all goroutines to complete
	em.wg.Wait()

	// Close channel only after all goroutines complete
	if em.keyEventChan != nil {
		close(em.keyEventChan)
	}

	em.isRunning = false
}

// getKeyEventChannel returns channel for receiving key  notifications
func (em *listenerKeyEventManager) getKeyEventChannel() <-chan KeyEvent {
	if em == nil {
		fmt.Printf("DEBUG: getKeyEventChannel called on nil manager\n")
		return nil
	}
	return em.keyEventChan
}

// getKeyValue tries to get the value of the key
func (em *listenerKeyEventManager) getKeyValue(key string) (string, error) {
	// Fast attempt to get the value with a short timeout
	ctx, cancel := context.WithTimeout(em.ctx, 5*time.Second)
	defer cancel()

	result, err := em.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key %s not found", key)
		}
		return "", fmt.Errorf("failed to get key value %s: %w", key, err)
	}

	return result, nil
}

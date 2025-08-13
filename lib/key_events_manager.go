package redisgklib

import (
	"context"
	"fmt"
	"strings"
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

	// Subscribe to specific Redis keyevent channels
	channels := []string{
		"__keyevent@0__:expire",  // TTL setting events
		"__keyevent@0__:expired", // Key expiration events
		"__keyevent@0__:set",     // Creation/update events
		"__keyevent@0__:del",     // Deletion events
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

	channelName := msg.Channel
	// Handle keyevent events
	if strings.HasPrefix(msg.Channel, "__keyevent@0__:") {
		key = msg.Payload
		// Determine event type from keyevent channel
		if strings.HasSuffix(msg.Channel, ":expire") {
			eventType = EventTypeExpire
		} else if strings.HasSuffix(msg.Channel, ":expired") {
			eventType = EventTypeExpired
		} else if strings.HasSuffix(msg.Channel, ":set") {
			eventType = EventTypeCreated
		} else if strings.HasSuffix(msg.Channel, ":del") {
			eventType = EventTypeDeleted
		} else {
			eventType = EventTypeUnknown
		}
	} else {
		// Unknown channel
		eventType = EventTypeUnknown
		key = msg.Payload
	}

	// Get key value if possible
	value := ""
	value, _ = em.getKeyValue(key)

	now := time.Now().UTC()

	return KeyEvent{
		Key:       key,
		Value:     value,
		EventType: eventType,
		Timestamp: now,
		Channel:   channelName,
	}
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

// getKeyEventChannel returns channel for receiving key event notifications
func (em *listenerKeyEventManager) getKeyEventChannel() <-chan KeyEvent {
	if em == nil {
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

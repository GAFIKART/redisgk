package redisgklib

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// expirationManager - manager for working with key expiration notifications
type expirationManager struct {
	client         *redis.Client
	ctx            context.Context
	cancel         context.CancelFunc
	expirationChan chan KeyExpirationEvent
	mu             sync.RWMutex
	isRunning      bool
	wg             sync.WaitGroup // Add WaitGroup for proper goroutine completion
}

// newExpirationManager creates a new key expiration notification manager
func newExpirationManager(client *redis.Client, ctx context.Context) *expirationManager {
	if client == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	managerCtx, cancel := context.WithCancel(ctx)

	return &expirationManager{
		client:         client,
		ctx:            managerCtx,
		cancel:         cancel,
		expirationChan: make(chan KeyExpirationEvent), // Unbuffered channel for simple forwarding
		isRunning:      false,
	}
}

// start starts the key expiration notification listener
func (em *expirationManager) start() error {
	if em == nil {
		return fmt.Errorf("expiration manager is nil")
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	if em.isRunning {
		// If listener is already running, just return success
		return nil
	}

	// Create subscription to key expiration notification channel
	pubsub := em.client.Subscribe(em.ctx, "__keyevent@0__:expire")

	// Start goroutine for processing notifications
	em.wg.Add(1)
	go em.listenForExpirations(pubsub)

	em.isRunning = true
	return nil
}

// listenForExpirations listens for key expiration notifications
func (em *expirationManager) listenForExpirations(pubsub *redis.PubSub) {
	defer func() {
		pubsub.Close()
		em.wg.Done()
	}()

	for {
		select {
		case <-em.ctx.Done():
			return
		case msg := <-pubsub.Channel():
			if msg.Channel == "__keyevent@0__:expire" {
				// Get record value before it expires
				value, err := em.getKeyValueBeforeExpiration(msg.Payload)
				if err != nil {
					// If failed to get value, use empty string
					value = ""
				}

				event := KeyExpirationEvent{
					Key:       msg.Payload,
					Value:     value,
					ExpiredAt: time.Now().UTC(),
				}


				// Simply forward event to user (block until user reads)
				select {
				case em.expirationChan <- event:
				case <-em.ctx.Done():
					return
				}
			}
		}
	}
}

// stop stops the notification listener
func (em *expirationManager) stop() {
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
	if em.expirationChan != nil {
		close(em.expirationChan)
	}

	em.isRunning = false
}

// getExpirationChannel returns channel for receiving key expiration notifications
func (em *expirationManager) getExpirationChannel() <-chan KeyExpirationEvent {
	if em == nil {
		fmt.Printf("DEBUG: getExpirationChannel called on nil manager\n")
		return nil
	}
	return em.expirationChan
}

// getKeyValueBeforeExpiration tries to get the value of the key before expiration
func (em *expirationManager) getKeyValueBeforeExpiration(key string) (string, error) {
	// Fast attempt to get the value with a short timeout
	ctx, cancel := context.WithTimeout(em.ctx, 50*time.Millisecond)
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

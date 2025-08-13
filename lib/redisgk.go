package redisgklib

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisGk - main structure for working with Redis
type RedisGk struct {
	redisClient *redis.Client
	baseCtx     time.Duration
	// Key notification manager
	listenerKeyEventManager *listenerKeyEventManager
}

// NewRedisGk creates a new RedisGk instance
func NewRedisGk(conf RedisConfConn) (*RedisGk, error) {
	// Check for empty configuration
	if (RedisConfConn{}) == conf {
		return nil, fmt.Errorf("configuration is empty")
	}

	if conf.AdditionalOptions.BaseCtx == 0 {
		conf.AdditionalOptions.BaseCtx = 10 * time.Second
	}

	redisClient, err := newRedisClientConnector(conf)
	if err != nil {
		return nil, err
	}

	// Create context for initialization
	ctx := context.Background()

	// Initialize Redis client with configuration check and subscription to notifications
	initializer := newRedisInitializer(redisClient, ctx)
	if initializer == nil {
		return nil, fmt.Errorf("failed to create redis initializer")
	}
	if err := initializer.initializeWithKeyExpirationNotifications(); err != nil {
		return nil, err
	}

	// Create key  notification manager
	listenerKeyEventManager := newListenerKeyEventManager(redisClient, context.Background())
	if listenerKeyEventManager == nil {
		return nil, fmt.Errorf("failed to create listener key event manager")
	}

	redisGk := &RedisGk{
		redisClient:             redisClient,
		baseCtx:                 conf.AdditionalOptions.BaseCtx,
		listenerKeyEventManager: listenerKeyEventManager,
	}

	// Automatically start key  notification listener
	if err := redisGk.listenerKeyEventManager.start(); err != nil {
		return nil, err
	}

	return redisGk, nil
}

// Close closes Redis connection
func (v *RedisGk) Close() error {
	// Stop notification manager
	if v.listenerKeyEventManager != nil {
		v.listenerKeyEventManager.stop()
	}

	if v.redisClient != nil {
		return v.redisClient.Close()
	}
	return nil
}

// ListenChannelKeyEventManager returns channel for receiving key  notifications
// Simple method for external library users
func (v *RedisGk) ListenChannelKeyEventManager() <-chan KeyEvent {
	if v == nil {
		return nil
	}
	if v.listenerKeyEventManager != nil {
		return v.listenerKeyEventManager.getKeyEventChannel()
	}
	return nil
}

// GetRedisClient returns the Redis client
func (v *RedisGk) GetRedisClient() *redis.Client {
	return v.redisClient
}

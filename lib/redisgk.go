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
	// Key expiration notification manager
	expirationManager *expirationManager
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

	// Create key expiration notification manager
	expirationManager := newExpirationManager(redisClient, context.Background())
	if expirationManager == nil {
		return nil, fmt.Errorf("failed to create expiration manager")
	}

	redisGk := &RedisGk{
		redisClient:       redisClient,
		baseCtx:           conf.AdditionalOptions.BaseCtx,
		expirationManager: expirationManager,
	}

	// Automatically start key expiration notification listener
	if err := redisGk.expirationManager.start(); err != nil {
		return nil, err
	}

	// Logging Redis server configuration at the end of connection
	fmt.Printf("📋 Redis Server Configuration:\n")

	// Getting all Redis configurations
	config, err := redisClient.ConfigGet(ctx, "*").Result()
	if err == nil {
		for name, value := range config {
			fmt.Printf("   %s: %s\n", name, value)
		}
	}

	return redisGk, nil
}

// Close closes Redis connection
func (v *RedisGk) Close() error {
	// Stop notification manager
	if v.expirationManager != nil {
		v.expirationManager.stop()
	}

	if v.redisClient != nil {
		return v.redisClient.Close()
	}
	return nil
}

// ListenChannelExpirationManager returns channel for receiving key expiration notifications
// Simple method for external library users
func (v *RedisGk) ListenChannelExpirationManager() <-chan KeyExpirationEvent {
	if v == nil {
		return nil
	}
	if v.expirationManager != nil {
		return v.expirationManager.getExpirationChannel()
	}
	return nil
}

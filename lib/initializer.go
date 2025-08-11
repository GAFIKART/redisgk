package redisgklib

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisInitializer - structure for Redis client initialization
type redisInitializer struct {
	client *redis.Client
	ctx    context.Context
}

// newRedisInitializer creates a new Redis initializer instance
func newRedisInitializer(client *redis.Client, ctx context.Context) *redisInitializer {
	if client == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	return &redisInitializer{
		client: client,
		ctx:    ctx,
	}
}

// initializeWithKeyExpirationNotifications initializes Redis client with subscription to key expiration notifications
func (ri *redisInitializer) initializeWithKeyExpirationNotifications() error {
	if ri == nil {
		return fmt.Errorf("redis initializer is nil")
	}

	// Check Redis connection
	if err := ri.checkConnection(); err != nil {
		return fmt.Errorf("error connecting to Redis: %w", err)
	}

	// Check and set configuration for key expiration notifications
	if err := ri.setupKeyExpirationNotifications(); err != nil {
		return fmt.Errorf("error setting up key expiration notifications: %w", err)
	}

	return nil
}

// checkConnection checks Redis connection
func (ri *redisInitializer) checkConnection() error {
	if ri.client == nil {
		return fmt.Errorf("error: Redis client is nil")
	}

	ctx, cancel := context.WithTimeout(ri.ctx, 5*time.Second)
	defer cancel()

	// Check ping
	if err := ri.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("error connecting to Redis: %w", err)
	}

	return nil
}

// setupKeyExpirationNotifications sets up key expiration notifications
func (ri *redisInitializer) setupKeyExpirationNotifications() error {
	if ri == nil {
		return fmt.Errorf("redis initializer is nil")
	}

	ctx, cancel := context.WithTimeout(ri.ctx, 5*time.Second)
	defer cancel()

	// Set configuration for key expiration notifications (Redis handles duplicates automatically)
	err := ri.client.ConfigSet(ctx, "notify-keyspace-events", "Exg").Err()
	if err != nil {
		return fmt.Errorf("error setting notify-keyspace-events: %w", err)
	}

	return nil
}

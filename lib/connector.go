package redisgklib

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// newRedisClientConnector creates a new Redis client
func newRedisClientConnector(conf RedisConfConn) (*redis.Client, error) {
	// Check for empty configuration
	if (RedisConfConn{}) == conf {
		return nil, fmt.Errorf("configuration is empty")
	}

	redisHost := conf.Host
	redisPort := conf.Port
	redisUser := conf.User
	redisPassword := conf.Password

	redisNDb := max(conf.DB, 0)

	if err := validateRedisConfConn(conf); err != nil {
		return nil, err
	}

	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
		Username: redisUser,
		Password: redisPassword,
		DB:       redisNDb,
	}

	opts = setRedisAdditionalOptions(opts, conf.AdditionalOptions)

	redisClient := redis.NewClient(opts)

	// Check Redis connection
	if err := testRedisConnection(redisClient); err != nil {
		return nil, fmt.Errorf("error: Redis connection error: %w", err)
	}

	return redisClient, nil
}

// testRedisConnection checks Redis connection
func testRedisConnection(client *redis.Client) error {
	if client == nil {
		return fmt.Errorf("error: Redis client is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check ping
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("error: Redis ping failed: %w", err)
	}

	return nil
}

// setRedisAdditionalOptions sets additional options for Redis client
func setRedisAdditionalOptions(opts *redis.Options, additionalOptions RedisAdditionalOptions) *redis.Options {
	if opts == nil {
		return nil
	}

	defaultDialTimeout := 10 * time.Second
	defaultReadTimeout := 30 * time.Second
	defaultWriteTimeout := 30 * time.Second
	defaultPoolSize := 20
	defaultPoolTimeout := 30 * time.Second

	if additionalOptions.DialTimeout != 0 {
		defaultDialTimeout = additionalOptions.DialTimeout
	}
	if additionalOptions.ReadTimeout != 0 {
		defaultReadTimeout = additionalOptions.ReadTimeout
	}
	if additionalOptions.WriteTimeout != 0 {
		defaultWriteTimeout = additionalOptions.WriteTimeout
	}
	if additionalOptions.PoolSize != 0 {
		defaultPoolSize = additionalOptions.PoolSize
	}
	if additionalOptions.PoolTimeout != 0 {
		defaultPoolTimeout = additionalOptions.PoolTimeout
	}

	opts.DialTimeout = defaultDialTimeout
	opts.ReadTimeout = defaultReadTimeout
	opts.WriteTimeout = defaultWriteTimeout
	opts.PoolSize = defaultPoolSize
	opts.PoolTimeout = defaultPoolTimeout

	return opts
}

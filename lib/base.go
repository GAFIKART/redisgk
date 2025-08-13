package redisgklib

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

// getKeyValue tries to get the value of the key
func (v *RedisGk) getKeyValue(key string) (string, error) {
	// Fast attempt to get the value with a short timeout
	ctx, cancel := v.createContextWithTimeout()
	defer cancel()

	result, err := v.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key %s not found", key)
		}
		return "", fmt.Errorf("failed to get key value %s: %w", key, err)
	}

	return result, nil
}

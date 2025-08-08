package redisgklib

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

// LPush adds elements to the beginning of the list
func (v *RedisGk) LPush(keyPath []string, values ...string) error {
	if v == nil {
		return fmt.Errorf("RedisGk instance is nil")
	}

	ctx, cancel := v.createContextWithTimeout()
	defer cancel()

	keyP, err := slicePathsConvertor(keyPath)
	if err != nil {
		return fmt.Errorf("key conversion error: %w", err)
	}

	// Check for empty values
	if len(values) == 0 {
		return fmt.Errorf("no values provided for LPush")
	}

	// Check for empty strings in values
	for i, value := range values {
		if value == "" {
			return fmt.Errorf("empty value at index %d", i)
		}
	}

	_, err = v.redisClient.LPush(ctx, keyP, values).Result()
	if err != nil {
		return fmt.Errorf("error adding to list: %w", err)
	}

	return nil
}

// RPush adds elements to the end of the list
func (v *RedisGk) RPush(keyPath []string, values ...string) error {
	if v == nil {
		return fmt.Errorf("RedisGk instance is nil")
	}

	ctx, cancel := v.createContextWithTimeout()
	defer cancel()

	keyP, err := slicePathsConvertor(keyPath)
	if err != nil {
		return fmt.Errorf("key conversion error: %w", err)
	}

	// Check for empty values
	if len(values) == 0 {
		return fmt.Errorf("no values provided for RPush")
	}

	// Check for empty strings in values
	for i, value := range values {
		if value == "" {
			return fmt.Errorf("empty value at index %d", i)
		}
	}

	_, err = v.redisClient.RPush(ctx, keyP, values).Result()
	if err != nil {
		return fmt.Errorf("error adding to list: %w", err)
	}

	return nil
}

// LPop removes and returns the first element of the list
func (v *RedisGk) LPop(keyPath []string) (string, error) {
	if v == nil {
		return "", fmt.Errorf("RedisGk instance is nil")
	}

	ctx, cancel := v.createContextWithTimeout()
	defer cancel()

	keyP, err := slicePathsConvertor(keyPath)
	if err != nil {
		return "", fmt.Errorf("key conversion error: %w", err)
	}

	result, err := v.redisClient.LPop(ctx, keyP).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("list is empty: %s", keyP)
		}
		return "", fmt.Errorf("error getting element from list: %w", err)
	}

	return result, nil
}

// RPop removes and returns the last element of the list
func (v *RedisGk) RPop(keyPath []string) (string, error) {
	if v == nil {
		return "", fmt.Errorf("RedisGk instance is nil")
	}

	ctx, cancel := v.createContextWithTimeout()
	defer cancel()

	keyP, err := slicePathsConvertor(keyPath)
	if err != nil {
		return "", fmt.Errorf("key conversion error: %w", err)
	}

	result, err := v.redisClient.RPop(ctx, keyP).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("list is empty: %s", keyP)
		}
		return "", fmt.Errorf("error getting element from list: %w", err)
	}

	return result, nil
}

// LRange returns list elements in the specified range
func (v *RedisGk) LRange(keyPath []string, start, stop int64) ([]string, error) {
	if v == nil {
		return nil, fmt.Errorf("RedisGk instance is nil")
	}

	ctx, cancel := v.createContextWithTimeout()
	defer cancel()

	keyP, err := slicePathsConvertor(keyPath)
	if err != nil {
		return nil, fmt.Errorf("key conversion error: %w", err)
	}

	result, err := v.redisClient.LRange(ctx, keyP, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("error getting list elements: %w", err)
	}

	return result, nil
}

// LLen returns the length of the list
func (v *RedisGk) LLen(keyPath []string) (int64, error) {
	if v == nil {
		return 0, fmt.Errorf("RedisGk instance is nil")
	}

	ctx, cancel := v.createContextWithTimeout()
	defer cancel()

	keyP, err := slicePathsConvertor(keyPath)
	if err != nil {
		return 0, fmt.Errorf("key conversion error: %w", err)
	}

	result, err := v.redisClient.LLen(ctx, keyP).Result()
	if err != nil {
		return 0, fmt.Errorf("error getting list length: %w", err)
	}

	return result, nil
}

package redisgklib

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// SetObj saves object to Redis with automatic JSON serialization
func SetObj[T any](
	v *RedisGk,
	keyPath []string,
	value T,
	ttlSlice ...time.Duration,
) error {
	if v == nil {
		return fmt.Errorf("RedisGk instance is nil")
	}

	ctx, cancel := v.createContextWithTimeout()
	defer cancel()

	keyP, err := slicePathsConvertor(keyPath)
	if err != nil {
		return fmt.Errorf("key conversion error: %w", err)
	}

	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("object serialization error: %w", err)
	}

	err = checkMaxSizeData(jsonData)
	if err != nil {
		return err
	}

	ttl := time.Duration(0)
	if len(ttlSlice) > 0 {
		ttl = ttlSlice[0]
	}

	return v.redisClient.Set(ctx, keyP, jsonData, ttl).Err()
}

// SetString saves string to Redis
func (v *RedisGk) SetString(
	keyPath []string,
	value string,
	ttlSlice ...time.Duration,
) error {
	if v == nil {
		return fmt.Errorf("RedisGk instance is nil")
	}

	ctx, cancel := v.createContextWithTimeout()
	defer cancel()

	keyP, err := slicePathsConvertor(keyPath)
	if err != nil {
		return fmt.Errorf("key conversion error: %w", err)
	}

	err = checkMaxSizeKey(keyP)
	if err != nil {
		return err
	}

	// Check value size
	if len(value) > maxSizeData {
		return fmt.Errorf("value size (%d bytes) exceeds Redis limit (512 MB)", len(value))
	}

	ttl := time.Duration(0)
	if len(ttlSlice) > 0 {
		ttl = ttlSlice[0]
	}

	return v.redisClient.Set(ctx, keyP, value, ttl).Err()
}

// GetObj gets object from Redis with automatic JSON deserialization
func GetObj[T any](
	v *RedisGk,
	keyPath []string,
) (*T, error) {
	if v == nil {
		return nil, fmt.Errorf("RedisGk instance is nil")
	}

	ctx, cancel := v.createContextWithTimeout()
	defer cancel()

	keyP, err := slicePathsConvertor(keyPath)
	if err != nil {
		return nil, fmt.Errorf("key conversion error: %w", err)
	}

	jsonStr, err := v.redisClient.Get(ctx, keyP).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("key not found: %s", keyP)
		}
		return nil, fmt.Errorf("error getting key %s: %w", keyP, err)
	}

	var result T
	err = json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, fmt.Errorf("object deserialization error: %w", err)
	}

	return &result, nil
}

// GetString gets string from Redis
func (v *RedisGk) GetString(
	keyPath []string,
) (string, error) {
	if v == nil {
		return "", fmt.Errorf("RedisGk instance is nil")
	}

	ctx, cancel := v.createContextWithTimeout()
	defer cancel()

	keyP, err := slicePathsConvertor(keyPath)
	if err != nil {
		return "", fmt.Errorf("key conversion error: %w", err)
	}

	result, err := v.redisClient.Get(ctx, keyP).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key not found: %s", keyP)
		}
		return "", fmt.Errorf("error getting key %s: %w", keyP, err)
	}

	return result, nil
}

// Del deletes one or multiple keys from Redis
func (v *RedisGk) Del(keyPath ...[]string) error {
	if v == nil {
		return fmt.Errorf("RedisGk instance is nil")
	}

	ctx, cancel := v.createContextWithTimeout()
	defer cancel()

	if len(keyPath) == 0 {
		return fmt.Errorf("no keys specified for deletion")
	}

	keysPDel := make([]string, 0, len(keyPath))
	for i, key := range keyPath {
		keyM, err := slicePathsConvertor(key)
		if err != nil {
			return fmt.Errorf("key conversion error %d: %w", i, err)
		}
		keysPDel = append(keysPDel, keyM)
	}

	result, err := v.redisClient.Del(ctx, keysPDel...).Result()
	if err != nil {
		return fmt.Errorf("error deleting keys: %w", err)
	}

	// Check that at least one key was deleted
	if result == 0 {
		return fmt.Errorf("none of the specified keys were found for deletion")
	}

	return nil
}

// FindKeyByPattern finds key by pattern and returns its value
func (v *RedisGk) FindKeyByPattern(patterns []string) (string, string, error) {
	if v == nil || v.redisClient == nil {
		return "", "", fmt.Errorf("listener key event manager or client is nil")
	}

	pattern := strings.Join(patterns, ":")
	pattern = pathRedisController(pattern)

	ctx, cancel := v.createContextWithTimeout()
	defer cancel()

	// Use SCAN to find keys by pattern
	iter := v.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		// Get key value
		value, err := v.getKeyValue(key)
		if err != nil {
			if err == redis.Nil {
				continue // Key already deleted
			}
			return key, "", fmt.Errorf("failed to get value for key %s: %w", key, err)
		}
		return key, value, nil // Return first found key and its value
	}

	if err := iter.Err(); err != nil {
		return "", "", fmt.Errorf("scan error: %w", err)
	}

	return "", "", fmt.Errorf("no keys found for pattern %s", pattern)
}

// FindObj searches objects by key pattern
func FindObj[T any](
	v *RedisGk,
	patternPath []string,
	countRes ...int64,
) (map[string]*T, error) {
	if v == nil {
		return nil, fmt.Errorf("RedisGk instance is nil")
	}

	ctx, cancel := v.createContextWithTimeout()
	defer cancel()

	pattern, err := slicePathsConvertor(patternPath)
	if err != nil {
		return nil, fmt.Errorf("pattern conversion error: %w", err)
	}
	pattern += "*"

	results := make(map[string]*T)
	var cursor uint64

	var count int64 = 100 // Default value
	if len(countRes) > 0 {
		count = countRes[0]
		if count <= 0 {
			count = 100
		}
	}

	// Process results directly without additional goroutines
	for {
		var keys []string
		keys, cursor, err = v.redisClient.Scan(ctx, cursor, pattern, count).Result()
		if err != nil {
			return nil, fmt.Errorf("key scanning error: %w", err)
		}

		if len(keys) == 0 {
			if cursor == 0 {
				break
			}
			continue
		}

		// Get values for all keys in one request
		values, err := v.redisClient.MGet(ctx, keys...).Result()
		if err != nil {
			return nil, fmt.Errorf("error getting values: %w", err)
		}

		// Process results
		for i, value := range values {
			if value == nil {
				continue
			}

			jsonStr, ok := value.(string)
			if !ok {
				continue
			}

			var obj T
			err = json.Unmarshal([]byte(jsonStr), &obj)
			if err != nil {
				// Skip objects with deserialization errors
				continue
			}

			// Add result directly to map
			results[keys[i]] = &obj
		}

		if cursor == 0 {
			break
		}
	}

	return results, nil
}

// GetKeys returns list of keys by pattern
func (v *RedisGk) GetKeys(patternPath []string) ([]string, error) {
	if v == nil {
		return nil, fmt.Errorf("RedisGk instance is nil")
	}

	ctx, cancel := v.createContextWithTimeout()
	defer cancel()

	pattern, err := slicePathsConvertor(patternPath)
	if err != nil {
		return nil, fmt.Errorf("pattern conversion error: %w", err)
	}
	pattern += "*"

	var allKeys []string
	var cursor uint64

	for {
		var keys []string
		keys, cursor, err = v.redisClient.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("key scanning error: %w", err)
		}

		allKeys = append(allKeys, keys...)

		if cursor == 0 {
			break
		}
	}

	return allKeys, nil
}

// Exists checks key existence
func (v *RedisGk) Exists(key []string) (bool, error) {
	if v == nil {
		return false, fmt.Errorf("RedisGk instance is nil")
	}

	ctx, cancel := v.createContextWithTimeout()
	defer cancel()

	keyP, err := slicePathsConvertor(key)
	if err != nil {
		return false, fmt.Errorf("key conversion error: %w", err)
	}

	result, err := v.redisClient.Exists(ctx, keyP).Result()
	if err != nil {
		return false, fmt.Errorf("error checking key existence: %w", err)
	}

	return result > 0, nil
}

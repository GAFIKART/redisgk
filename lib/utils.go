package redisgklib

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// createContextWithTimeout creates context with timeout for Redis operations
func (v *RedisGk) createContextWithTimeout() (context.Context, context.CancelFunc) {
	if v == nil {
		// Return context with default timeout if instance is nil
		return context.WithTimeout(context.Background(), 10*time.Second)
	}
	return context.WithTimeout(context.Background(), v.baseCtx)
}

// pathRedisController normalizes key for Redis
func pathRedisController(key string) string {
	if key == "" {
		return ""
	}

	keys := strings.ToLower(key)

	// Fix regular expression - remove extra characters
	re01 := regexp.MustCompile(`[\*\?\[\]\.]`)
	keys = re01.ReplaceAllString(keys, "")

	// Replace multiple colons with single one
	re02 := regexp.MustCompile(`:{2,}`)
	keys = re02.ReplaceAllString(keys, ":")

	// Replace spaces with underscores
	keys = strings.ReplaceAll(keys, " ", "_")

	// Remove colons at beginning and end
	keys = strings.Trim(keys, ":")

	// Check for maximum key length
	if len(keys) > maxSizeData {
		// Truncate key to maximum length
		keys = keys[:maxSizeData]
	}

	return keys
}

// slicePathsConvertor converts string slice to Redis key path
func slicePathsConvertor(keySlice []string) (string, error) {
	if keySlice == nil {
		return "", fmt.Errorf("keySlice is nil")
	}

	if len(keySlice) == 0 {
		return "", fmt.Errorf("keySlice is empty")
	}

	// Check each slice element
	for i, key := range keySlice {
		if key == "" {
			return "", fmt.Errorf("element %d in keySlice is empty", i)
		}
	}

	keyPath := strings.Join(keySlice, ":")
	keyPath = pathRedisController(keyPath)

	// Check result after normalization
	if keyPath == "" {
		return "", fmt.Errorf("key normalization result is empty")
	}

	err := checkMaxSizeKey(keyPath)
	if err != nil {
		return "", err
	}

	return keyPath, nil
}

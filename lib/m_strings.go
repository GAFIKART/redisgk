package redisgklib

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func SetObj[T any](
	v *RedisGk,
	keyPath []string,
	value T,
	ttlSlice ...time.Duration,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), v.baseCtx)
	defer cancel()

	keyP, err := slicePathsConvertor(keyPath)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
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

func (v *RedisGk) SetString(
	keyPath []string,
	value string,
	ttlSlice ...time.Duration,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), v.baseCtx)
	defer cancel()

	keyP, err := slicePathsConvertor(keyPath)
	if err != nil {
		return err
	}

	err = checkMaxSizeKey(keyP)
	if err != nil {
		return err
	}

	ttl := time.Duration(0)
	if len(ttlSlice) > 0 {
		ttl = ttlSlice[0]
	}

	return v.redisClient.Set(ctx, keyP, value, ttl).Err()
}

func GetObj[T any](
	v *RedisGk,
	keyPath []string,
) (*T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), v.baseCtx)
	defer cancel()
	keyP, err := slicePathsConvertor(keyPath)
	if err != nil {
		return nil, err
	}
	jsonStr, err := v.redisClient.Get(ctx, keyP).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("key not found: %s", keyP)
		}
		return nil, err
	}
	var result T
	err = json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (v *RedisGk) GetString(
	keyPath []string,
) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), v.baseCtx)
	defer cancel()

	keyP, err := slicePathsConvertor(keyPath)
	if err != nil {
		return "", err
	}

	result, err := v.redisClient.Get(ctx, keyP).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key not found: %s", keyP)
		}
		return "", err
	}

	return result, nil
}

func (v *RedisGk) Del(keyPath ...[]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), v.baseCtx)
	defer cancel()

	
	keysPDel := make([]string, 0)
	for _, key := range keyPath {
		keyM, err := slicePathsConvertor(key)
		if err != nil {
			return err
		}
		keysPDel = append(keysPDel, keyM)
	}

	return v.redisClient.Del(ctx, keysPDel...).Err()
}

func FindObj[T any](
	v *RedisGk,
	patternPath []string,
	countRes ...int64,
) (map[string]*T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), v.baseCtx)
	defer cancel()

	pattern, err := slicePathsConvertor(patternPath)
	if err != nil {
		return nil, err
	}
	pattern += "*"

	results := make(map[string]*T)
	var cursor uint64

	var count int64
	if len(countRes) > 0 {
		count = countRes[0]
	}

	for {
		var keys []string
		keys, cursor, err = v.redisClient.Scan(ctx, cursor, pattern, count).Result()
		if err != nil {
			return nil, fmt.Errorf("error scan keys: %w", err)
		}

		if len(keys) == 0 {
			if cursor == 0 {
				break
			}
			continue
		}

		values, err := v.redisClient.MGet(ctx, keys...).Result()
		if err != nil {
			return nil, fmt.Errorf("error mget values: %w", err)
		}

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
				continue
			}

			results[keys[i]] = &obj
		}

		if cursor == 0 {
			break
		}
	}

	return results, nil
}

func (v *RedisGk) Exists(key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), v.baseCtx)
	defer cancel()

	key = pathRedisController(key)

	result, err := v.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

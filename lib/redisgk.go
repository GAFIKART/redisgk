package redisgklib

import (
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisGk - основная структура для работы с Redis
type RedisGk struct {
	redisClient *redis.Client
	baseCtx     time.Duration
}

// NewRedisGk создает новый экземпляр RedisGk
func NewRedisGk(conf RedisConfConn) (*RedisGk, error) {

	if conf.AdditionalOptions.BaseCtx == 0 {
		conf.AdditionalOptions.BaseCtx = 10 * time.Second
	}

	redisClient, err := newRedisClientConnector(conf)
	if err != nil {
		return nil, err
	}

	return &RedisGk{
		redisClient: redisClient,
		baseCtx:     conf.AdditionalOptions.BaseCtx,
	}, nil
}

// Close закрывает соединение с Redis
func (v *RedisGk) Close() error {
	if v.redisClient != nil {
		return v.redisClient.Close()
	}
	return nil
}

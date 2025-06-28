package redisgklib

import (
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisOnce   sync.Once
	redisClient *redis.Client
)

func newRedisClientConnector(conf RedisConfConn) (*redis.Client, error) {
	var err error
	redisOnce.Do(func() {
		redisHost := conf.Host
		redisPort := conf.Port
		redisUser := conf.User
		redisPassword := conf.Password

		redisNDb := max(conf.DB, 0)

		if err = validateRedisConfConn(conf); err != nil {
			return
		}

		opts := &redis.Options{
			Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
			Username: redisUser,
			Password: redisPassword,
			DB:       redisNDb,
		}

		opts = setRedisAdditionalOptions(opts, conf.AdditionalOptions)

		redisClient = redis.NewClient(opts)
	})
	return redisClient, err
}

func setRedisAdditionalOptions(opts *redis.Options, additionalOptions RedisAdditionalOptions) *redis.Options {

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

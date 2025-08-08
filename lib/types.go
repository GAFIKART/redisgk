package redisgklib

import (
	"time"
)

type RedisConfConn struct {
	Host     string
	Port     int
	User     string
	Password string
	DB       int

	AdditionalOptions RedisAdditionalOptions
}

type RedisAdditionalOptions struct {
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	PoolTimeout  time.Duration

	BaseCtx time.Duration
}

// KeyExpirationEvent - structure for key expiration event
type KeyExpirationEvent struct {
	Key       string    `json:"key"`        // Key name
	Value     string    `json:"value"`      // Record body (value)
	ExpiredAt time.Time `json:"expired_at"` // Expiration time
}

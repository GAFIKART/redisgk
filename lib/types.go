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

// EventType - Redis event type
type EventType string

const (
	EventTypeExpired EventType = "expired" // Key expired
	EventTypeCreated EventType = "created" // Key created
	EventTypeUpdated EventType = "updated" // Key updated
	EventTypeDeleted EventType = "deleted" // Key deleted
	EventTypeUnknown EventType = "unknown" // Unknown event type
)

// KeyEvent - structure for Redis key event
type KeyEvent struct {
	Key       string    `json:"key"`        // Key name
	Value     string    `json:"value"`      // Record body (value)
	EventType EventType `json:"event_type"` // Event type
	Timestamp time.Time `json:"timestamp"`  // Event timestamp
}

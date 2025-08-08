package tests

import (
	"testing"
	"time"

	redisgklib "github.com/GAFIKART/redisgk/lib"
)

// TestValidateRedisConfConn tests configuration validation
func TestValidateRedisConfConn(t *testing.T) {
	tests := []struct {
		config   redisgklib.RedisConfConn
		hasError bool
	}{
		{
			redisgklib.RedisConfConn{
				Host:     "localhost",
				Port:     6379,
				Password: "password",
				DB:       0,
			},
			false,
		},
		{
			redisgklib.RedisConfConn{
				Host:     "",
				Port:     6379,
				Password: "password",
				DB:       0,
			},
			true,
		},
		{
			redisgklib.RedisConfConn{
				Host:     "localhost",
				Port:     0,
				Password: "password",
				DB:       0,
			},
			true,
		},
		{
			redisgklib.RedisConfConn{
				Host:     "localhost",
				Port:     6379,
				Password: "",
				DB:       0,
			},
			true,
		},
	}

	for _, test := range tests {
		_, err := redisgklib.NewRedisGk(test.config)
		if test.hasError && err == nil {
			t.Errorf("NewRedisGk should return error for %+v", test.config)
		}
		if !test.hasError && err != nil {
			t.Errorf("NewRedisGk returned error for %+v: %v", test.config, err)
		}
	}
}

// TestKeyExpirationEventStructure tests key expiration event structure
func TestKeyExpirationEventStructure(t *testing.T) {
	event := redisgklib.KeyExpirationEvent{
		Key:       "test:key",
		Value:     "test value",
		ExpiredAt: time.Now(),
	}

	if event.Key != "test:key" {
		t.Errorf("KeyExpirationEvent.Key = %q, expected %q", event.Key, "test:key")
	}

	if event.Value != "test value" {
		t.Errorf("KeyExpirationEvent.Value = %q, expected %q", event.Value, "test value")
	}
}

// TestRedisConfConnStructure tests RedisConfConn structure
func TestRedisConfConnStructure(t *testing.T) {
	config := redisgklib.RedisConfConn{
		Host:     "localhost",
		Port:     6379,
		User:     "user",
		Password: "password",
		DB:       0,
		AdditionalOptions: redisgklib.RedisAdditionalOptions{
			DialTimeout:  10 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			PoolSize:     20,
			PoolTimeout:  30 * time.Second,
			BaseCtx:      10 * time.Second,
		},
	}

	if config.Host != "localhost" {
		t.Errorf("Expected Host to be 'localhost', got %s", config.Host)
	}

	if config.Port != 6379 {
		t.Errorf("Expected Port to be 6379, got %d", config.Port)
	}

	if config.User != "user" {
		t.Errorf("Expected User to be 'user', got %s", config.User)
	}

	if config.Password != "password" {
		t.Errorf("Expected Password to be 'password', got %s", config.Password)
	}

	if config.DB != 0 {
		t.Errorf("Expected DB to be 0, got %d", config.DB)
	}
}

// TestRedisAdditionalOptionsStructure tests RedisAdditionalOptions structure
func TestRedisAdditionalOptionsStructure(t *testing.T) {
	options := redisgklib.RedisAdditionalOptions{
		DialTimeout:  5 * time.Second,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		PoolSize:     10,
		PoolTimeout:  20 * time.Second,
		BaseCtx:      5 * time.Second,
	}

	if options.DialTimeout != 5*time.Second {
		t.Errorf("Expected DialTimeout to be 5s, got %v", options.DialTimeout)
	}

	if options.ReadTimeout != 15*time.Second {
		t.Errorf("Expected ReadTimeout to be 15s, got %v", options.ReadTimeout)
	}

	if options.WriteTimeout != 15*time.Second {
		t.Errorf("Expected WriteTimeout to be 15s, got %v", options.WriteTimeout)
	}

	if options.PoolSize != 10 {
		t.Errorf("Expected PoolSize to be 10, got %d", options.PoolSize)
	}

	if options.PoolTimeout != 20*time.Second {
		t.Errorf("Expected PoolTimeout to be 20s, got %v", options.PoolTimeout)
	}

	if options.BaseCtx != 5*time.Second {
		t.Errorf("Expected BaseCtx to be 5s, got %v", options.BaseCtx)
	}
}

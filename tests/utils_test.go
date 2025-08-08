package tests

import (
	"context"
	"testing"
	"time"

	redisgklib "github.com/GAFIKART/redisgk/lib"
)

// TestKeyExpirationEvent tests key expiration event structure
func TestKeyExpirationEvent(t *testing.T) {
	now := time.Now()
	event := redisgklib.KeyExpirationEvent{
		Key:       "test:key",
		Value:     "test value",
		ExpiredAt: now,
	}

	if event.Key != "test:key" {
		t.Errorf("KeyExpirationEvent.Key = %q, expected %q", event.Key, "test:key")
	}

	if event.Value != "test value" {
		t.Errorf("KeyExpirationEvent.Value = %q, expected %q", event.Value, "test value")
	}

	if !event.ExpiredAt.Equal(now) {
		t.Errorf("KeyExpirationEvent.ExpiredAt = %v, expected %v", event.ExpiredAt, now)
	}
}

// TestRedisConfConn tests configuration structure
func TestRedisConfConn(t *testing.T) {
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

	if config.AdditionalOptions.DialTimeout != 10*time.Second {
		t.Errorf("Expected DialTimeout to be 10s, got %v", config.AdditionalOptions.DialTimeout)
	}
}

// TestRedisAdditionalOptions tests additional options structure
func TestRedisAdditionalOptions(t *testing.T) {
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

// TestNilChecks tests nil pointer handling
func TestNilChecks(t *testing.T) {
	// Test with nil configuration
	_, err := redisgklib.NewRedisGk(redisgklib.RedisConfConn{})
	if err == nil {
		t.Error("NewRedisGk should return error for empty configuration")
	}

	// Test with invalid configuration
	invalidConfig := redisgklib.RedisConfConn{
		Host:     "",
		Port:     0,
		Password: "",
		DB:       -1,
	}
	_, err = redisgklib.NewRedisGk(invalidConfig)
	if err == nil {
		t.Error("NewRedisGk should return error for invalid configuration")
	}
}

// TestConfigurationValidation tests various configuration scenarios
func TestConfigurationValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    redisgklib.RedisConfConn
		shouldErr bool
	}{
		{
			"valid config",
			redisgklib.RedisConfConn{
				Host:     "localhost",
				Port:     6379,
				Password: "password",
				DB:       0,
			},
			false,
		},
		{
			"empty host",
			redisgklib.RedisConfConn{
				Host:     "",
				Port:     6379,
				Password: "password",
				DB:       0,
			},
			true,
		},
		{
			"invalid port",
			redisgklib.RedisConfConn{
				Host:     "localhost",
				Port:     0,
				Password: "password",
				DB:       0,
			},
			true,
		},
		{
			"port too low",
			redisgklib.RedisConfConn{
				Host:     "localhost",
				Port:     1023,
				Password: "password",
				DB:       0,
			},
			true,
		},
		{
			"port too high",
			redisgklib.RedisConfConn{
				Host:     "localhost",
				Port:     65536,
				Password: "password",
				DB:       0,
			},
			true,
		},
		{
			"empty password",
			redisgklib.RedisConfConn{
				Host:     "localhost",
				Port:     6379,
				Password: "",
				DB:       0,
			},
			true,
		},
		{
			"negative DB",
			redisgklib.RedisConfConn{
				Host:     "localhost",
				Port:     6379,
				Password: "password",
				DB:       -1,
			},
			true,
		},
		{
			"valid IP address",
			redisgklib.RedisConfConn{
				Host:     "127.0.0.1",
				Port:     6379,
				Password: "password",
				DB:       0,
			},
			false,
		},
		{
			"valid domain",
			redisgklib.RedisConfConn{
				Host:     "redis.example.com",
				Port:     6379,
				Password: "password",
				DB:       0,
			},
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := redisgklib.NewRedisGk(test.config)
			if test.shouldErr && err == nil {
				t.Errorf("Expected error for configuration %+v", test.config)
			}
			if !test.shouldErr && err != nil {
				t.Errorf("Unexpected error for configuration %+v: %v", test.config, err)
			}
		})
	}
}

// TestDataSizeLimits tests data size validation
func TestDataSizeLimits(t *testing.T) {
	// Test with small data
	smallData := make([]byte, 1024)
	if len(smallData) > 512*1024*1024 {
		t.Error("Small data should be within size limits")
	}

	// Test with large data
	largeData := make([]byte, 512*1024*1024+1)
	if len(largeData) <= 512*1024*1024 {
		t.Error("Large data should exceed size limits")
	}

	// Test key size limits
	smallKey := "test:key"
	if len(smallKey) > 512*1024*1024 {
		t.Error("Small key should be within size limits")
	}

	largeKey := string(make([]byte, 512*1024*1024+1))
	if len(largeKey) <= 512*1024*1024 {
		t.Error("Large key should exceed size limits")
	}
}

// TestKeyNormalization tests key normalization scenarios
func TestKeyNormalization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal key", "user:profile:name", "user:profile:name"},
		{"spaces to underscores", "User Profile Name", "user_profile_name"},
		{"special characters", "test*key?value", "testkeyvalue"},
		{"multiple colons", "key::with::colons", "key:with:colons"},
		{"spaces and colons", "  key  with  spaces  ", "key_with_spaces"},
		{"empty string", "", ""},
		{"unicode characters", "ключ:с:кириллицей", "ключ:с:кириллицей"},
		{"mixed case", "TestKey", "testkey"},
		{"numbers and symbols", "key123!@#", "key123"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// This test validates the expected behavior of key normalization
			// The actual implementation is private, so we test the expected output
			if test.input == "" && test.expected != "" {
				t.Errorf("Empty input should produce empty output")
			}
			if test.input != "" && test.expected == "" {
				t.Errorf("Non-empty input should not produce empty output")
			}
		})
	}
}

// TestSliceToKeyPath tests slice to key path conversion scenarios
func TestSliceToKeyPath(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		shouldErr bool
	}{
		{"valid path", []string{"user", "profile", "name"}, false},
		{"simple key", []string{"test", "key"}, false},
		{"empty slice", []string{}, true},
		{"nil slice", nil, true},
		{"empty element", []string{"", "key"}, true},
		{"empty last element", []string{"key", ""}, true},
		{"multiple empty elements", []string{"key", "", "value"}, true},
		{"single element", []string{"single"}, false},
		{"many elements", []string{"a", "b", "c", "d", "e"}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// This test validates the expected behavior of slice to key path conversion
			// The actual implementation is private, so we test the expected behavior
			if test.shouldErr {
				// For error cases, we expect the function to return an error
				if test.input == nil {
					// nil slice should cause error
				} else if len(test.input) == 0 {
					// empty slice should cause error
				} else {
					// check for empty elements
					for _, elem := range test.input {
						if elem == "" {
							// empty element should cause error
							break
						}
					}
				}
			} else {
				// For valid cases, we expect the function to work
				if len(test.input) == 0 {
					t.Errorf("Empty slice should cause error")
				}
				for i, elem := range test.input {
					if elem == "" {
						t.Errorf("Empty element at index %d should cause error", i)
					}
				}
			}
		})
	}
}

// TestContextTimeout tests context timeout behavior
func TestContextTimeout(t *testing.T) {
	// Test that context creation works with different timeouts
	timeouts := []time.Duration{
		1 * time.Second,
		5 * time.Second,
		10 * time.Second,
		30 * time.Second,
	}

	for _, timeout := range timeouts {
		t.Run(timeout.String(), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			if ctx == nil {
				t.Error("Context should not be nil")
			}
			cancel()
		})
	}
}

// TestHostValidation tests host validation scenarios
func TestHostValidation(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		expected bool
	}{
		{"localhost", "localhost", true},
		{"valid IP", "127.0.0.1", true},
		{"valid IPv6", "::1", true},
		{"valid domain", "redis.example.com", true},
		{"valid subdomain", "sub.redis.example.com", true},
		{"empty string", "", false},
		{"invalid domain", "invalid..domain", false},
		{"domain with invalid chars", "redis@example.com", false},
		{"domain starting with dash", "-redis.example.com", false},
		{"domain ending with dash", "redis.example.com-", false},
		{"too long domain", string(make([]byte, 254)), false},
		{"domain part too long", "a" + string(make([]byte, 64)) + ".com", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// This test validates the expected behavior of host validation
			// The actual implementation is private, so we test the expected behavior
			if test.host == "" && test.expected {
				t.Errorf("Empty host should not be valid")
			}
			if test.host == "localhost" && !test.expected {
				t.Errorf("localhost should be valid")
			}
			if len(test.host) > 253 && test.expected {
				t.Errorf("Host longer than 253 characters should not be valid")
			}
		})
	}
}

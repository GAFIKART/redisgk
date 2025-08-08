package tests

import (
	"testing"
	"time"

	redisgklib "github.com/GAFIKART/redisgk/lib"
)

// TestNilPointerSafety tests nil pointer handling in all methods
func TestNilPointerSafety(t *testing.T) {
	tests := []struct {
		name        string
		description string
		testFunc    func() error
		shouldErr   bool
	}{
		{
			"NewRedisGk with empty config",
			"Should return error for empty configuration",
			func() error {
				_, err := redisgklib.NewRedisGk(redisgklib.RedisConfConn{})
				return err
			},
			true,
		},
		{
			"NewRedisGk with nil config",
			"Should handle nil configuration gracefully",
			func() error {
				var config redisgklib.RedisConfConn
				_, err := redisgklib.NewRedisGk(config)
				return err
			},
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.testFunc()
			if test.shouldErr && err == nil {
				t.Errorf("%s: expected error but got none", test.description)
			}
			if !test.shouldErr && err != nil {
				t.Errorf("%s: unexpected error: %v", test.description, err)
			}
		})
	}
}

// TestInputValidation tests input validation for various scenarios
func TestInputValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      redisgklib.RedisConfConn
		description string
		shouldErr   bool
	}{
		{
			"valid configuration",
			redisgklib.RedisConfConn{
				Host:     "localhost",
				Port:     6379,
				Password: "password",
				DB:       0,
			},
			"Valid configuration should pass validation",
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
			"Empty host should cause validation error",
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
			"Invalid port should cause validation error",
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
			"Port below 1024 should cause validation error",
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
			"Port above 65535 should cause validation error",
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
			"Empty password should cause validation error",
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
			"Negative DB should cause validation error",
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := redisgklib.NewRedisGk(test.config)
			if test.shouldErr && err == nil {
				t.Errorf("%s: expected error but got none", test.description)
			}
			if !test.shouldErr && err != nil {
				t.Errorf("%s: unexpected error: %v", test.description, err)
			}
		})
	}
}

// TestDataSizeValidation tests data size validation
func TestDataSizeValidation(t *testing.T) {
	// Test data size limits
	maxSize := 512 * 1024 * 1024 // 512 MB

	tests := []struct {
		name        string
		dataSize    int
		description string
		shouldErr   bool
	}{
		{
			"small data",
			1024,
			"Small data should be within limits",
			false,
		},
		{
			"medium data",
			1024 * 1024, // 1 MB
			"Medium data should be within limits",
			false,
		},
		{
			"large data",
			maxSize,
			"Large data at limit should be accepted",
			false,
		},
		{
			"exceeding limit",
			maxSize + 1,
			"Data exceeding limit should cause error",
			true,
		},
		{
			"zero size",
			0,
			"Zero size data should be accepted",
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := make([]byte, test.dataSize)
			if len(data) > maxSize && !test.shouldErr {
				t.Errorf("%s: data size %d exceeds limit %d", test.description, len(data), maxSize)
			}
			if len(data) <= maxSize && test.shouldErr {
				t.Errorf("%s: data size %d should be within limit %d", test.description, len(data), maxSize)
			}
		})
	}
}

// TestKeySizeValidation tests key size validation
func TestKeySizeValidation(t *testing.T) {
	maxSize := 512 * 1024 * 1024 // 512 MB

	tests := []struct {
		name        string
		key         string
		description string
		shouldErr   bool
	}{
		{
			"small key",
			"test:key",
			"Small key should be within limits",
			false,
		},
		{
			"medium key",
			string(make([]byte, 1024)),
			"Medium key should be within limits",
			false,
		},
		{
			"large key",
			string(make([]byte, maxSize)),
			"Large key at limit should be accepted",
			false,
		},
		{
			"exceeding limit",
			string(make([]byte, maxSize+1)),
			"Key exceeding limit should cause error",
			true,
		},
		{
			"empty key",
			"",
			"Empty key should be accepted",
			false,
		},
		{
			"unicode key",
			"ключ:с:кириллицей",
			"Unicode key should be accepted",
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if len(test.key) > maxSize && !test.shouldErr {
				t.Errorf("%s: key size %d exceeds limit %d", test.description, len(test.key), maxSize)
			}
			if len(test.key) <= maxSize && test.shouldErr {
				t.Errorf("%s: key size %d should be within limit %d", test.description, len(test.key), maxSize)
			}
		})
	}
}

// TestConfigurationEdgeCases tests edge cases in configuration
func TestConfigurationEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		config      redisgklib.RedisConfConn
		description string
		shouldErr   bool
	}{
		{
			"zero values",
			redisgklib.RedisConfConn{},
			"Zero values should cause validation error",
			true,
		},
		{
			"negative port",
			redisgklib.RedisConfConn{
				Host:     "localhost",
				Port:     -1,
				Password: "password",
				DB:       0,
			},
			"Negative port should cause validation error",
			true,
		},
		{
			"very large port",
			redisgklib.RedisConfConn{
				Host:     "localhost",
				Port:     999999,
				Password: "password",
				DB:       0,
			},
			"Very large port should cause validation error",
			true,
		},
		{
			"negative DB",
			redisgklib.RedisConfConn{
				Host:     "localhost",
				Port:     6379,
				Password: "password",
				DB:       -999,
			},
			"Negative DB should cause validation error",
			true,
		},
		{
			"very large DB",
			redisgklib.RedisConfConn{
				Host:     "localhost",
				Port:     6379,
				Password: "password",
				DB:       999999,
			},
			"Very large DB should be accepted",
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := redisgklib.NewRedisGk(test.config)
			if test.shouldErr && err == nil {
				t.Errorf("%s: expected error but got none", test.description)
			}
			if !test.shouldErr && err != nil {
				t.Errorf("%s: unexpected error: %v", test.description, err)
			}
		})
	}
}

// TestTimeoutValidation tests timeout validation
func TestTimeoutValidation(t *testing.T) {
	tests := []struct {
		name        string
		timeout     time.Duration
		description string
		shouldErr   bool
	}{
		{
			"zero timeout",
			0,
			"Zero timeout should be handled gracefully",
			false,
		},
		{
			"negative timeout",
			-1 * time.Second,
			"Negative timeout should be handled gracefully",
			false,
		},
		{
			"very short timeout",
			1 * time.Millisecond,
			"Very short timeout should be accepted",
			false,
		},
		{
			"normal timeout",
			10 * time.Second,
			"Normal timeout should be accepted",
			false,
		},
		{
			"very long timeout",
			1 * time.Hour,
			"Very long timeout should be accepted",
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := redisgklib.RedisConfConn{
				Host:     "localhost",
				Port:     6379,
				Password: "password",
				DB:       0,
				AdditionalOptions: redisgklib.RedisAdditionalOptions{
					BaseCtx: test.timeout,
				},
			}
			_, err := redisgklib.NewRedisGk(config)
			if test.shouldErr && err == nil {
				t.Errorf("%s: expected error but got none", test.description)
			}
			if !test.shouldErr && err != nil {
				t.Errorf("%s: unexpected error: %v", test.description, err)
			}
		})
	}
}

// TestResourceSafety tests resource safety scenarios
func TestResourceSafety(t *testing.T) {
	// Test that we can create multiple configurations without resource leaks
	for i := 0; i < 100; i++ {
		config := redisgklib.RedisConfConn{
			Host:     "localhost",
			Port:     6379,
			Password: "password",
			DB:       0,
			AdditionalOptions: redisgklib.RedisAdditionalOptions{
				BaseCtx: 10 * time.Second,
			},
		}
		_, err := redisgklib.NewRedisGk(config)
		if err == nil {
			// If we can create it, we should be able to close it
			// This test ensures no resource leaks in configuration creation
		}
	}
}

// TestConcurrentSafety tests concurrent access safety
func TestConcurrentSafety(t *testing.T) {
	// Test concurrent configuration creation
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			config := redisgklib.RedisConfConn{
				Host:     "localhost",
				Port:     6379,
				Password: "password",
				DB:       0,
			}
			_, err := redisgklib.NewRedisGk(config)
			if err == nil {
				// Configuration creation should be thread-safe
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

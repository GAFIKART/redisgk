package redisgklib

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// validateRedisConfConn validates Redis connection configuration
func validateRedisConfConn(conf RedisConfConn) error {
	if conf.Host == "" {
		return errors.New("host is required")
	}

	// Check that host is a valid IP or domain name
	if !isValidHost(conf.Host) {
		return fmt.Errorf("invalid host: %s", conf.Host)
	}

	if conf.Port == 0 {
		return errors.New("port is required")
	}
	if conf.Port < 1 || conf.Port > 65535 {
		return fmt.Errorf("port must be in range 1-65535, got: %d", conf.Port)
	}
	if conf.Port < 1024 {
		return errors.New("port must be >= 1024 (privileged ports require additional permissions)")
	}

	if conf.Password == "" {
		return errors.New("password is required")
	}

	if conf.DB < 0 {
		return fmt.Errorf("DB must be >= 0, got: %d", conf.DB)
	}

	return nil
}

// isValidHost checks if host is valid
func isValidHost(host string) bool {
	// Check that it's not an empty string
	if host == "" {
		return false
	}

	// Check that it's not localhost or IP address
	if host == "localhost" || net.ParseIP(host) != nil {
		return true
	}

	// Check that it's a valid domain name
	if len(host) > 253 {
		return false
	}

	// Simple domain name validation
	for part := range strings.SplitSeq(host, ".") {
		if len(part) == 0 || len(part) > 63 {
			return false
		}
		// Check for valid characters in domain name
		for _, char := range part {
			if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
				return false
			}
		}
		// Check that domain doesn't start or end with hyphen
		if len(part) > 0 && (part[0] == '-' || part[len(part)-1] == '-') {
			return false
		}
	}

	return true
}

const maxSizeData = int(512 * 1024 * 1024) // 512 MB

// checkMaxSizeData checks data size
func checkMaxSizeData(data []byte) error {
	if len(data) > maxSizeData {
		return fmt.Errorf("data size (%d bytes) exceeds Redis limit (512 MB)", len(data))
	}
	return nil
}

// checkMaxSizeKey checks key size
func checkMaxSizeKey(key string) error {
	if len(key) > maxSizeData {
		return fmt.Errorf("key size (%d bytes) exceeds Redis limit (512 MB)", len(key))
	}
	return nil
}

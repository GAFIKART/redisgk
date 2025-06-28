package redisgklib

import (
	"errors"
	"fmt"
)

func validateRedisConfConn(conf RedisConfConn) error {
	if conf.Host == "" {
		return errors.New("host is required")
	}
	if conf.Port == 0 {
		return errors.New("port is required")
	}
	if conf.Port < 1024 {
		return errors.New("port must be >= 1024 (privileged ports require additional permissions)")
	}

	if conf.Password == "" {
		return errors.New("password is required")
	}

	return nil
}

const maxSizeData = int(512 * 1024 * 1024)

func checkMaxSizeData(data []byte) error {
	if len(data) > maxSizeData {
		return fmt.Errorf("size data (%d bytes) exceeds the Redis limit (512 MB)", len(data))
	}
	return nil
}

func checkMaxSizeKey(key string) error {
	if len(key) > maxSizeData {
		return fmt.Errorf("size key (%d bytes) exceeds the Redis limit (512 MB)", len(key))
	}
	return nil
}

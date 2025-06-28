package redisgklib

import (
	"fmt"
	"regexp"
	"strings"
)

func pathRedisController(key string) string {
	keys := strings.ToLower(key)
	re01 := regexp.MustCompile(`[\*\?\]\[\.]`)
	keys = re01.ReplaceAllString(keys, "")

	re02 := regexp.MustCompile(`:{2,}`)
	keys = re02.ReplaceAllString(keys, ":")

	keys = strings.ReplaceAll(keys, " ", "_")

	keys = strings.Trim(keys, ":")

	return keys
}

func slicePathsConvertor(keySlice []string) (string, error) {
	if len(keySlice) == 0 {
		return "", fmt.Errorf("keySlice is empty")
	}

	keyPath := strings.Join(keySlice, ":")
	keyPath = pathRedisController(keyPath)

	err := checkMaxSizeKey(keyPath)
	if err != nil {
		return "", err
	}

	return keyPath, nil
}

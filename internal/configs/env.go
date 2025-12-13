package configs

import (
	"os"
	"strconv"
)

func getEnv(k, def string) string {
	v := os.Getenv(k)
	if v != "" {
		return v
	}
	return def
}

func getInt(k string, def int) int {
	v := os.Getenv(k)
	if v != "" {
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
	}
	return def
}

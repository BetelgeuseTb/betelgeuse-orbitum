package cache

import "time"

type Cache interface {
	Get(key string) (string, bool, error)
	Set(key, value string, ttl time.Duration) error
	Del(key string) error
}

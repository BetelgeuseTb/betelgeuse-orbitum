package cache

import (
	"context"
	"errors"
	"time"
)

var (
	ErrCacheMiss = errors.New("cache miss")
	ErrClosed    = errors.New("cache manager is closed")
)

type Cache interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type Manager interface {
	Cache(name string) Cache
	Shutdown(ctx context.Context) error
}

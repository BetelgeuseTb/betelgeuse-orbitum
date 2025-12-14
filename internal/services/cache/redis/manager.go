package redis

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/services/cache"
	"github.com/redis/go-redis/v9"
)

type Manager struct {
	client     *redis.Client
	defaultTTL time.Duration

	mu     sync.RWMutex
	caches map[string]*Cache
	closed atomic.Bool
}

type Option func(*Manager)

func WithDefaultTTL(ttl time.Duration) Option {
	return func(m *Manager) {
		m.defaultTTL = ttl
	}
}

func NewManager(client *redis.Client, opts ...Option) *Manager {
	m := &Manager{
		client:     client,
		defaultTTL: 5 * time.Minute,
		caches:     make(map[string]*Cache),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Manager) Cache(name string) cache.Cache {
	if m.closed.Load() {
		return &errorCache{err: cache.ErrClosed}
	}

	m.mu.RLock()
	c, exists := m.caches[name]
	m.mu.RUnlock()

	if exists {
		return c
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if c, exists = m.caches[name]; exists {
		return c
	}

	c = newCache(m.client, name, m.defaultTTL)
	m.caches[name] = c
	return c
}

func (m *Manager) Shutdown(ctx context.Context) error {
	if !m.closed.CompareAndSwap(false, true) {
		return nil
	}

	return m.client.Close()
}

type errorCache struct {
	err error
}

func (e *errorCache) Get(context.Context, string, any) error {
	return e.err
}

func (e *errorCache) Set(context.Context, string, any, time.Duration) error {
	return e.err
}

func (e *errorCache) Delete(context.Context, string) error {
	return e.err
}

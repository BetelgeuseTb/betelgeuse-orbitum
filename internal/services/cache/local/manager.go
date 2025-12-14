package local

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/services/cache"
)

type Manager struct {
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

func NewManager(opts ...Option) *Manager {
	m := &Manager{
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

	c = newCache(name, m.defaultTTL)
	m.caches[name] = c
	return c
}

func (m *Manager) GetAllCaches() []*Cache {
	m.mu.RLock()
	defer m.mu.RUnlock()

	caches := make([]*Cache, 0, len(m.caches))
	for _, c := range m.caches {
		caches = append(caches, c)
	}
	return caches
}

func (m *Manager) Shutdown(ctx context.Context) error {
	if !m.closed.CompareAndSwap(false, true) {
		return nil
	}
	return nil
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

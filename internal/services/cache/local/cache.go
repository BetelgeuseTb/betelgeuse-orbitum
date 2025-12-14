package local

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/services/cache"
)

type Cache struct {
	name       string
	defaultTTL time.Duration
	mu         sync.RWMutex
	items      map[string]item
}

type item struct {
	value     []byte
	expiresAt time.Time
}

func newCache(name string, defaultTTL time.Duration) *Cache {
	return &Cache{
		name:       name,
		defaultTTL: defaultTTL,
		items:      make(map[string]item),
	}
}

func (c *Cache) Get(_ context.Context, key string, dest any) error {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return cache.ErrCacheMiss
	}

	if time.Now().After(item.expiresAt) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return cache.ErrCacheMiss
	}

	return json.Unmarshal(item.value, dest)
}

func (c *Cache) Set(_ context.Context, key string, value any, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.items[key] = item{
		value:     data,
		expiresAt: time.Now().Add(ttl),
	}
	c.mu.Unlock()

	return nil
}

func (c *Cache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
	return nil
}

func (c *Cache) EvictExpired() int {
	now := time.Now()
	evicted := 0

	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.items {
		if now.After(item.expiresAt) {
			delete(c.items, key)
			evicted++
		}
	}

	return evicted
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *Cache) GetCacheName() string {
	return c.name
}

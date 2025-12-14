package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/services/cache"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client     *redis.Client
	namespace  string
	defaultTTL time.Duration
}

func newCache(client *redis.Client, namespace string, defaultTTL time.Duration) *Cache {
	return &Cache{
		client:     client,
		namespace:  namespace,
		defaultTTL: defaultTTL,
	}
}

func (c *Cache) buildKey(key string) string {
	return fmt.Sprintf("%s:%s", c.namespace, key)
}

func (c *Cache) Get(ctx context.Context, key string, dest any) error {
	data, err := c.client.Get(ctx, c.buildKey(key)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return cache.ErrCacheMiss
		}
		return err
	}

	return json.Unmarshal(data, dest)
}

func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, c.buildKey(key), data, ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, c.buildKey(key)).Err()
	if err == redis.Nil {
		return nil
	}
	return err
}

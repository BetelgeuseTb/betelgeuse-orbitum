package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	c   *redis.Client
	ctx context.Context
}

func NewRedisCache(c *redis.Client) *RedisCache {
	return &RedisCache{c: c, ctx: context.Background()}
}

func (r *RedisCache) Get(key string) (string, bool, error) {
	val, err := r.c.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

func (r *RedisCache) Set(key, value string, ttl time.Duration) error {
	return r.c.Set(r.ctx, key, value, ttl).Err()
}

func (r *RedisCache) Del(key string) error {
	return r.c.Del(r.ctx, key).Err()
}

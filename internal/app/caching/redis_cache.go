package caching

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

// RedisCache структура для реализации интерфейса Cache с использованием Redis.
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache конструктор для создания экземпляра RedisCache.
func NewRedisCache(addr string) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr, // Адрес сервера Redis.
	})
	return &RedisCache{
		client: rdb,
		ctx:    context.Background(),
	}
}

// Set реализация метода Set интерфейса Cache.
func (cache *RedisCache) Set(key string, value any, expiration time.Duration) error {
	return cache.client.Set(cache.ctx, key, value, expiration).Err()
}

// Get реализация метода Get интерфейса Cache.
func (cache *RedisCache) Get(key string) (any, error) {
	return cache.client.Get(cache.ctx, key).Result()
}

// Delete реализация метода Delete интерфейса Cache.
func (cache *RedisCache) Delete(key string) error {
	return cache.client.Del(cache.ctx, key).Err()
}

// Purge реализация метода Purge интерфейса Cache.
// Redis не имеет прямого метода для очистки всего ключевого пространства,
// поэтому мы используем команду FLUSHDB.
func (cache *RedisCache) Purge() error {
	return cache.client.FlushDB(cache.ctx).Err()
}

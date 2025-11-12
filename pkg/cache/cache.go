// Package cache provides a lightweight interface for caching structured
// data using Redis. It supports JSON serialization for convenience and
// defines a generic Cache interface that can be implemented by other backends.
package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache defines the generic cache operations used by services.
// Values are stored as JSON and retrieved into arbitrary structs.
type Cache interface {
	// GetJSON retrieves a cached item into out.
	// It returns (false, nil) if the key is not found.
	GetJSON(ctx context.Context, key string, out any) (bool, error)

	// SetJSON stores an object with a given TTL (in seconds).
	// If ttl <= 0, the default TTL is used.
	SetJSON(ctx context.Context, key string, v any, ttl time.Duration) error

	// Delete removes an item from the cache.
	Delete(ctx context.Context, key string) error

	// DefaultTTL returns the default time-to-live for cached items.
	DefaultTTL() time.Duration
}

// redisCache implements Cache using Redis as the backend.
type redisCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisCache returns a new Redis-backed Cache instance.
//
// Example:
//
//	client := redis.NewClient(&redis.Options{
//	    Addr: "localhost:6379",
//	})
//	cache := cache.NewRedisCache(client, 10*time.Minute)
func NewRedisCache(client *redis.Client, defaultTTL time.Duration) Cache {
	return &redisCache{client: client, ttl: defaultTTL}
}

// DefaultTTL implements Cache.DefaultTTL.
func (c *redisCache) DefaultTTL() time.Duration { return c.ttl }

// GetJSON implements Cache.GetJSON.
func (c *redisCache) GetJSON(ctx context.Context, key string, out any) (bool, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal([]byte(val), out)
}

// SetJSON implements Cache.SetJSON.
func (c *redisCache) SetJSON(ctx context.Context, key string, v any, ttl time.Duration) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if ttl <= 0 {
		ttl = c.ttl
	}
	return c.client.Set(ctx, key, data, ttl).Err()
}

// Delete implements Cache.Delete.
func (c *redisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

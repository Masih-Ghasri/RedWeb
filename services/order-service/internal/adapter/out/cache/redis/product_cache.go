package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// productCacheValue is the single wire format of a cached product. Both the
// writer (RedisProductCache) and the reader (RedisInventoryCache) decode into
// this type, which is what guarantees the two never drift apart.
type productCacheValue struct {
	Stock int     `json:"stock"`
	Price float64 `json:"price"`
}

// productKey returns the Redis key for a cached product. Exported nowhere by
// design: every read and write in this package must go through it.
func productKey(productID string) string {
	return "product:" + productID
}

// RedisProductCache is the write side of the shared product cache, fed by the
// product events catalog-service publishes.
type RedisProductCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisProductCache instantiates the Redis product cache writer.
func NewRedisProductCache(client *redis.Client, ttl time.Duration) *RedisProductCache {
	return &RedisProductCache{client: client, ttl: ttl}
}

func (c *RedisProductCache) SetProduct(ctx context.Context, productID string, stock int, price float64) error {
	payload, err := json.Marshal(productCacheValue{Stock: stock, Price: price})
	if err != nil {
		return fmt.Errorf("encode product cache entry for %s: %w", productID, err)
	}

	return c.client.Set(ctx, productKey(productID), payload, c.ttl).Err()
}

func (c *RedisProductCache) DeleteProduct(ctx context.Context, productID string) error {
	return c.client.Del(ctx, productKey(productID)).Err()
}

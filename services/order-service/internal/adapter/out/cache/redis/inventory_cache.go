package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/domain"
	"github.com/redis/go-redis/v9"
)

// RedisInventoryCache is the read side of the shared product cache. The write
// side is RedisProductCache in the same package; both use productKey and
// productCacheValue so the contract cannot be broken from one end only.
type RedisInventoryCache struct {
	client *redis.Client
}

func NewRedisInventoryCache(client *redis.Client) *RedisInventoryCache {
	return &RedisInventoryCache{client: client}
}

// GetProductStockAndPrice returns the cached stock and price for a product.
//
// A cache miss becomes domain.ErrProductNotFound (→404) while an
// infrastructure failure propagates unwrapped (→500), so "no such product"
// and "Redis is down" are never mistaken for each other.
func (c *RedisInventoryCache) GetProductStockAndPrice(ctx context.Context, productID string) (int, float64, error) {
	val, err := c.client.Get(ctx, productKey(productID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, 0, fmt.Errorf("%w: %s", domain.ErrProductNotFound, productID)
		}
		return 0, 0, fmt.Errorf("read product cache for %s: %w", productID, err)
	}

	var data productCacheValue
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return 0, 0, fmt.Errorf("decode product cache for %s: %w", productID, err)
	}

	return data.Stock, data.Price, nil
}

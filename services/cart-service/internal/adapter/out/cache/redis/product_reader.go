package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/core/domain"
	"github.com/redis/go-redis/v9"
)

// productCacheValue mirrors the JSON stored under "product:{id}".
//
// *** WIRE CONTRACT — READ BEFORE CHANGING ***
//
// The key format and this struct are duplicated verbatim from
// order-service/internal/adapter/out/cache/redis/product_cache.go, which is
// the writer fed by catalog-service's product events. cart-service cannot
// import that package (separate Go module), so the contract is restated here
// by hand. Change the tags on one side only and every cart operation starts
// failing with "product not found".
type productCacheValue struct {
	Stock int     `json:"stock"`
	Price float64 `json:"price"`
}

// RedisProductReader is the read side of that shared cache.
type RedisProductReader struct {
	client *redis.Client
}

func NewRedisProductReader(client *redis.Client) *RedisProductReader {
	return &RedisProductReader{client: client}
}

// GetProductInfo returns the cached stock and price for a product.
func (r *RedisProductReader) GetProductInfo(ctx context.Context, productID string) (int, float64, error) {
	val, err := r.client.Get(ctx, "product:"+productID).Result()
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

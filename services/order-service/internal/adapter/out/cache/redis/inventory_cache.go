package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisInventoryCache struct {
	client *redis.Client
}

func NewRedisInventoryCache(client *redis.Client) *RedisInventoryCache {
	return &RedisInventoryCache{client: client}
}

func (c *RedisInventoryCache) GetProductStockAndPrice(ctx context.Context, productID string) (int, float64, error) {
	val, err := c.client.Get(ctx, fmt.Sprintf("product:%s", productID)).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, 0, fmt.Errorf("product not found in cache")
		}
		return 0, 0, err
	}

	var data struct {
		Stock int     `json:"stock"`
		Price float64 `json:"price"`
	}

	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return 0, 0, err
	}

	return data.Stock, data.Price, nil
}

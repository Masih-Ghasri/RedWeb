package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisIdempotencyRepository struct {
	client *redis.Client
}

func NewRedisIdempotencyRepository(client *redis.Client) *RedisIdempotencyRepository {
	return &RedisIdempotencyRepository{client: client}
}

func (r *RedisIdempotencyRepository) LockIfNotExists(ctx context.Context, key string) (bool, error) {
	return r.client.SetNX(ctx, key, "locked", 24*time.Hour).Result()
}

func (r *RedisIdempotencyRepository) Unlock(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

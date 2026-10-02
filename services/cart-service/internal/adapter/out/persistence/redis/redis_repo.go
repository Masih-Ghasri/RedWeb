package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/core/domain"
	"github.com/redis/go-redis/v9"
)

// RedisCartRepository stores carts as a single JSON blob per customer.
//
// The blob (rather than a hash of line items) is deliberate: every mutation
// rewrites the whole cart anyway, there is no partial update to preserve,
// and one key means one expiry to reason about.
type RedisCartRepository struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisCartRepository builds the repository. ttl is applied on every Save,
// so an abandoned cart is swept without needing a background job.
func NewRedisCartRepository(client *redis.Client, ttl time.Duration) *RedisCartRepository {
	return &RedisCartRepository{
		client: client,
		ttl:    ttl,
	}
}

// cartKey is the single place the cart key format is spelled out.
func cartKey(customerID string) string {
	return "cart:" + customerID
}

func (r *RedisCartRepository) Get(ctx context.Context, customerID string) (*domain.Cart, error) {
	val, err := r.client.Get(ctx, cartKey(customerID)).Result()
	if err != nil {
		// A missing key is an empty cart, not a failure.
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("read cart for %s: %w", customerID, err)
	}

	var cart domain.Cart
	if err := json.Unmarshal([]byte(val), &cart); err != nil {
		return nil, fmt.Errorf("decode cart for %s: %w", customerID, err)
	}

	return &cart, nil
}

func (r *RedisCartRepository) Save(ctx context.Context, cart *domain.Cart) error {
	payload, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("encode cart for %s: %w", cart.CustomerID, err)
	}

	if err := r.client.Set(ctx, cartKey(cart.CustomerID), payload, r.ttl).Err(); err != nil {
		return fmt.Errorf("write cart for %s: %w", cart.CustomerID, err)
	}

	return nil
}

// Delete removes a cart. Redis reports zero affected keys for an absent one
// without erroring, which is what makes clearing twice safe.
func (r *RedisCartRepository) Delete(ctx context.Context, customerID string) error {
	if err := r.client.Del(ctx, cartKey(customerID)).Err(); err != nil {
		return fmt.Errorf("delete cart for %s: %w", customerID, err)
	}
	return nil
}

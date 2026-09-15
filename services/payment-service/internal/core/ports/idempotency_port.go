package ports

import "context"

type IdempotencyRepository interface {
	LockIfNotExists(ctx context.Context, key string) (bool, error)
	Unlock(ctx context.Context, key string) error
}

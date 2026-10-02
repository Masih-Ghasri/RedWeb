package ports

import (
	"context"

	"github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/core/domain"
)

type CartRepository interface {
	// Get returns (nil, nil) when the customer has no cart yet.
	//
	// Absence is not an error: "no cart" and "cart exists but is empty" are
	// the same state as far as the caller is concerned, and the service turns
	// both into an empty cart.
	Get(ctx context.Context, customerID string) (*domain.Cart, error)

	// Save writes the whole cart and refreshes its expiry.
	Save(ctx context.Context, cart *domain.Cart) error

	// Delete removes a cart. Deleting one that does not exist succeeds, so
	// CLEAR is idempotent.
	Delete(ctx context.Context, customerID string) error
}

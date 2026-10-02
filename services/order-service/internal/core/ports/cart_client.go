package ports

import (
	"context"
	"errors"
)

// ErrCartNotFound reports that cart-service knows nothing about the cart —
// distinct from ErrCartUnavailable (→503) so "your cart id is wrong" and
// "cart-service is down" never collapse into the same response.
var ErrCartNotFound = errors.New("cart not found")

// ErrCartUnavailable reports that cart-service could not be reached or gave
// an unusable answer. Checkout must fail closed here: building an order from
// a cart we could not read would be guessing.
var ErrCartUnavailable = errors.New("cart service unavailable")

// CartItemRef is one line of a cart as order-service needs it: enough to
// price and stock-check the line, nothing more.
type CartItemRef struct {
	ProductID string
	Quantity  int
}

// CartSnapshot is the checkout view of a cart.
//
// CustomerID comes from here rather than from the request body, which is what
// closes the IDOR on order creation — identity travels with the cart instead
// of being asserted by whoever is calling.
type CartSnapshot struct {
	CustomerID string
	Items      []CartItemRef
}

// CartClient is order-service's door to cart-service.
//
// It is named a client rather than a reader because checkout both reads a
// cart and, on success, clears it — a port called "Reader" with a mutating
// method would be a lie.
//
// GetCart is called synchronously during checkout because the order needs
// the cart *now*; an eventually-consistent cache read would let two
// checkouts race on the same stale snapshot. ClearCart is best-effort: a
// failure is logged and must never fail an order that is already committed.
type CartClient interface {
	GetCart(ctx context.Context, cartID string) (*CartSnapshot, error)
	ClearCart(ctx context.Context, cartID string) error
}

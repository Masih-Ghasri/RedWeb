package domain

import "errors"

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInsufficientStock  = errors.New("insufficient stock for product")
	ErrInvalidOrderAmount = errors.New("invalid total order amount")
	ErrEmptyOrderItems    = errors.New("order must contain at least one item")
	ErrOrderFailed        = errors.New("order failed")

	// ErrProductNotFound → 404. The product cache has no entry for an id, so
	// there is nothing to price the line with. Shared with cart-service's
	// reader, which reports the same condition for the same reason.
	ErrProductNotFound = errors.New("product not found")

	// ErrOrderNotCancelable → 409. The order exists but is no longer in a
	// state cancellation is meaningful for — typically payment already moved
	// it to PAID/FAILED, or another cancel won the race.
	ErrOrderNotCancelable = errors.New("order is not cancelable")
)

// Cart-service failures live on the port rather than here — see
// internal/core/ports/cart_client.go for ErrCartNotFound and
// ErrCartUnavailable. They describe conditions at a boundary, not facts about
// the order domain, and having two identically-worded sentinels in different
// packages would make errors.Is silently fail on the wrong one.

package domain

import "errors"

// Sentinel errors for the cart domain. The HTTP adapter maps each to a
// status code — see internal/adapter/in/http/errors.go — so the messages
// must describe their own condition rather than being generic.
var (
	// ErrProductNotFound → 404. Raised when the shared product cache has no
	// entry for a product id: the cart will not accept something it cannot
	// price. It also covers a genuinely unknown id, because the cache is the
	// only product source this service reads.
	ErrProductNotFound = errors.New("product not found")

	// ErrProductNotInCart → 404. The addressed cart line does not exist.
	ErrProductNotInCart = errors.New("product not in cart")

	// ErrInvalidQuantity → 400. Quantities must be whole positive numbers;
	// zero is only meaningful on an update, where it means "remove".
	ErrInvalidQuantity = errors.New("quantity must be greater than zero")

	// ErrInsufficientStock → 409. The request is well-formed but conflicts
	// with what is currently available.
	ErrInsufficientStock = errors.New("insufficient stock for product")
)

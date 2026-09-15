package domain

import "errors"

// Sentinel errors for the catalog domain.
//
// The HTTP adapter maps these to status codes — see
// internal/adapter/in/http/errors.go — so each message must describe its own
// condition precisely rather than being a generic "invalid input".
var (
	// ErrProductNotFound → 404
	ErrProductNotFound = errors.New("product not found")

	// ErrInvalidPrice → 400
	ErrInvalidPrice = errors.New("product price must be greater than zero")

	// ErrStockRequired → 400 (a product cannot be created without stock)
	ErrStockRequired = errors.New("product stock must be greater than zero")

	// ErrInvalidStock → 400 (stock may be updated to 0, but never below)
	ErrInvalidStock = errors.New("product stock cannot be negative")

	// ErrInvalidQuantity → 400
	ErrInvalidQuantity = errors.New("quantity must be greater than zero")

	// ErrInsufficientStock → 409 (the request is well-formed but conflicts
	// with the current stock level)
	ErrInsufficientStock = errors.New("insufficient stock for product")
)

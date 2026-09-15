package domain

import "errors"

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInsufficientStock  = errors.New("insufficient stock for product")
	ErrInvalidOrderAmount = errors.New("invalid total order amount")
	ErrEmptyOrderItems    = errors.New("order must contain at least one item")
)

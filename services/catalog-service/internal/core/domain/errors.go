package domain

import "errors"

var (
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidPrice    = errors.New("product price cannot be negative")
	ErrInvalidStock    = errors.New("product stock cannot be negative")
)

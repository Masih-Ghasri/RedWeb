package domain

import "errors"

var (
	ErrInvalidOrderID = errors.New("order ID cannot be empty")
	ErrShipmentFailed = errors.New("failed to process shipment")
)

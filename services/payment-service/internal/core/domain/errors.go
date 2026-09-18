package domain

import "errors"

var (
	ErrInvalidAmount      = errors.New("payment amount must be greater than zero")
	ErrPaymentAlreadyDone = errors.New("payment has already been processed")
)

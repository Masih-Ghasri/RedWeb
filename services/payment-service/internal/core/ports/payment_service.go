package ports

import "context"

type ProcessPaymentCommand struct {
	OrderID string
	Amount  float64
}

type PaymentUseCase interface {
	ProcessPayment(ctx context.Context, cmd ProcessPaymentCommand) error
}

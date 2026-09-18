package ports

import "context"

type PaymentGateway interface {
	Charge(ctx context.Context, orderID string, amount float64) (bool, error)
}

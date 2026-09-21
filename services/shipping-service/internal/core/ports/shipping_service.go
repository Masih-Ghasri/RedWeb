package ports

import "context"

type ProcessShipmentCommand struct {
	OrderID string
}

type ShippingUseCase interface {
	ProcessShipment(ctx context.Context, cmd ProcessShipmentCommand) error
}

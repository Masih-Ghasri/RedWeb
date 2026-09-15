package ports

import (
	"context"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/domain"
)

type CreateOrderItemCommand struct {
	ProductID string
	Quantity  int
}

type CreateOrderCommand struct {
	CustomerID string
	Items      []CreateOrderItemCommand
}

type OrderUseCase interface {
	CreateOrder(ctx context.Context, cmd CreateOrderCommand) (*domain.Order, error)
}

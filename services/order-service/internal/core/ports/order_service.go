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
	GetOrder(ctx context.Context, id string) (*domain.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status string) error
}

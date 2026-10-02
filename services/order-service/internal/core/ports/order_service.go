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
	// CartID is the primary identifier for creating orders from a shopping cart.
	CartID string

	// Fallback fields for direct order placement without cart.
	CustomerID string
	Items      []CreateOrderItemCommand
}

type ListOrdersQuery struct {
	CustomerID string
	Status     string
	Page       int
	PageSize   int
}

type OrderList struct {
	Items    []*domain.Order
	Page     int
	PageSize int
	Total    int
}

type OrderUseCase interface {
	CreateOrder(ctx context.Context, cmd CreateOrderCommand) (*domain.Order, error)
	GetOrder(ctx context.Context, id string) (*domain.Order, error)
	ListOrders(ctx context.Context, q ListOrdersQuery) (*OrderList, error)
	// CancelOrder moves a still-pending order to CANCELED, or returns
	// domain.ErrOrderNotCancelable if payment already claimed it.
	CancelOrder(ctx context.Context, id string) (*domain.Order, error)
	// UpdateOrderStatus is the payment consumer's entry point; it is not
	// exposed over HTTP.
	UpdateOrderStatus(ctx context.Context, orderID string, status string) error
}

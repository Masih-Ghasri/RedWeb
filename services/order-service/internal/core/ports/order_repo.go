package ports

import (
	"context"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/domain"
)

type OrderRepository interface {
	CreateOrderWithOutbox(ctx context.Context, order *domain.Order, event *domain.OutboxEvent) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
}

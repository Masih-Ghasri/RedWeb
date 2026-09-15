package ports

import (
	"context"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/domain"
)

type OutboxRepository interface {
	GetPendingEvents(ctx context.Context, limit int) ([]*domain.OutboxEvent, error)

	MarkAsPublished(ctx context.Context, id string) error
}

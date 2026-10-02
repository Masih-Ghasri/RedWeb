package ports

import (
	"context"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/domain"
)

// OrderFilter narrows and pages an order listing. Zero values mean "no
// restriction"; the service defaults Page/PageSize before querying.
type OrderFilter struct {
	CustomerID string
	Status     string
	Page       int // 1-based
	PageSize   int
}

type OrderRepository interface {
	// CreateOrderWithOutbox writes the order, its line items and its outbox
	// event in a single transaction — either all three land or none do.
	CreateOrderWithOutbox(ctx context.Context, order *domain.Order, event *domain.OutboxEvent) error

	// GetByID loads an order together with its items, or returns
	// domain.ErrOrderNotFound.
	GetByID(ctx context.Context, id string) (*domain.Order, error)

	// FindAll returns a page of orders plus the total matching the filter.
	FindAll(ctx context.Context, filter OrderFilter) ([]*domain.Order, int, error)

	// UpdateStatus changes an order's status and reports whether a row was
	// actually updated. A false result means the order does not exist —
	// callers must not treat it as success.
	UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) (bool, error)

	// CancelIfPending atomically moves PENDING → CANCELED and reports whether
	// it happened.
	//
	// The status guard is the point: without it, cancel could read PENDING,
	// lose the race to a payment that flips the row to PAID, and then
	// overwrite it — leaving a paid order looking cancelled. Returning false
	// lets the caller answer 409 instead of pretending to succeed.
	CancelIfPending(ctx context.Context, id string) (bool, error)
}

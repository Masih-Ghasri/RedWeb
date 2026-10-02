package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/ports"
)

type PostgresOutboxRepository struct {
	db *gorm.DB
}

func NewPostgresOutboxRepository(db *gorm.DB) *PostgresOutboxRepository {
	return &PostgresOutboxRepository{db: db}
}

// GetPendingEvents atomically claims outbox rows waiting to be published.
// Uses FOR UPDATE SKIP LOCKED inside a transaction to prevent duplicate delivery across replicas.
func (r *PostgresOutboxRepository) GetPendingEvents(ctx context.Context, limit int) ([]*domain.OutboxEvent, error) {
	var models []OutboxEventModel

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ?", string(domain.OutboxStatusPending)).
			Order("created_at ASC").
			Limit(limit).
			Find(&models).Error; err != nil {
			return err
		}

		if len(models) == 0 {
			return nil
		}

		ids := make([]string, len(models))
		for i, m := range models {
			ids[i] = m.ID
		}

		return tx.Model(&OutboxEventModel{}).
			Where("id IN ?", ids).
			Update("status", string(domain.OutboxStatusProcessing)).Error
	})
	if err != nil {
		return nil, err
	}

	var events []*domain.OutboxEvent
	for _, m := range models {
		events = append(events, &domain.OutboxEvent{
			ID:            m.ID,
			AggregateType: m.AggregateType,
			AggregateID:   m.AggregateID,
			EventType:     m.EventType,
			Payload:       m.Payload,
			Status:        domain.OutboxStatus(m.Status),
			CreatedAt:     m.CreatedAt,
		})
	}

	return events, nil
}

func (r *PostgresOutboxRepository) MarkAsPublished(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&OutboxEventModel{}).
		Where("id = ?", id).
		Update("status", string(domain.OutboxStatusPublished)).Error
}

type PostgresOrderRepository struct {
	db *gorm.DB
}

func NewPostgresOrderRepository(db *gorm.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

// CreateOrderWithOutbox writes the order, its line items and the outbox event
// in one transaction. GORM cascades the Items association because of the
// OrderID foreign key, so all three tables commit together.
func (r *PostgresOrderRepository) CreateOrderWithOutbox(ctx context.Context, order *domain.Order, event *domain.OutboxEvent) error {
	orderModel := toOrderModel(order)
	outboxModel := toOutboxModel(event)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(orderModel).Error; err != nil {
			return err
		}
		if err := tx.Create(outboxModel).Error; err != nil {
			return err
		}
		return nil
	})
}

// GetByID loads an order with its line items.
//
// This used to be a stub returning `nil, nil`, which is the worst possible
// shape: no error *and* no order, so the first caller to wire a GET route
// dereferenced nil. It now reports ErrOrderNotFound explicitly.
func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	var model OrderModel

	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("id = ?", id).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}

	return toDomainOrder(&model), nil
}

// FindAll returns one page of orders plus the total matching the filter.
func (r *PostgresOrderRepository) FindAll(ctx context.Context, filter ports.OrderFilter) ([]*domain.Order, int, error) {
	var total int64
	if err := r.query(ctx, filter).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, pageSize := normalizePagination(filter)

	var models []OrderModel
	if err := r.query(ctx, filter).
		Preload("Items").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&models).Error; err != nil {
		return nil, 0, err
	}

	orders := make([]*domain.Order, 0, len(models))
	for i := range models {
		orders = append(orders, toDomainOrder(&models[i]))
	}

	return orders, int(total), nil
}

// query builds a *fresh* statement on every call, so the Count and the Find
// in FindAll never share a GORM chain — reusing one would leave the second
// query with the first one's SELECT.
func (r *PostgresOrderRepository) query(ctx context.Context, filter ports.OrderFilter) *gorm.DB {
	q := r.db.WithContext(ctx).Model(&OrderModel{})

	if filter.CustomerID != "" {
		q = q.Where("customer_id = ?", filter.CustomerID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}

	return q
}

// UpdateStatus changes an order's status, reporting whether a row matched.
//
// Returning the affected count instead of swallowing it is deliberate: a
// false result means the order does not exist, and the payment consumer used
// to treat that as success and silently drop the transition.
func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) (bool, error) {
	res := r.db.WithContext(ctx).
		Model(&OrderModel{}).
		Where("id = ?", id).
		Update("status", string(status))
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// CancelIfPending performs the PENDING → CANCELED transition atomically.
//
// The `status = 'PENDING'` predicate is load-bearing: cancel reads the order
// first to produce a friendly error, and between that read and this write the
// payment consumer can flip the row to PAID. Guarding the UPDATE means the
// loser of that race affects zero rows and the caller can answer 409 instead
// of overwriting a paid order.
func (r *PostgresOrderRepository) CancelIfPending(ctx context.Context, id string) (bool, error) {
	res := r.db.WithContext(ctx).
		Model(&OrderModel{}).
		Where("id = ?", id).
		Where("status = ?", string(domain.StatusPending)).
		Update("status", string(domain.StatusCanceled))
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// toDomainOrder maps a stored order (and its eagerly-loaded items) back to
// the domain shape the use cases work with.
func toDomainOrder(m *OrderModel) *domain.Order {
	order := &domain.Order{
		ID:          m.ID,
		CustomerID:  m.CustomerID,
		TotalAmount: m.TotalAmount,
		Status:      domain.OrderStatus(m.Status),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}

	for _, item := range m.Items {
		order.Items = append(order.Items, domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	return order
}

// normalizePagination guarantees a usable pair even if a filter arrives
// unclamped — a negative page would otherwise produce a negative OFFSET.
func normalizePagination(filter ports.OrderFilter) (int, int) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = domain.DefaultPageSize
	}
	return page, pageSize
}

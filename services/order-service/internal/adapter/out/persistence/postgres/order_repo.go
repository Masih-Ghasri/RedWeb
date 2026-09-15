package postgres

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/domain"
)

type PostgresOutboxRepository struct {
	db *gorm.DB
}

func NewPostgresOutboxRepository(db *gorm.DB) *PostgresOutboxRepository {
	return &PostgresOutboxRepository{db: db}
}

func (r *PostgresOutboxRepository) GetPendingEvents(ctx context.Context, limit int) ([]*domain.OutboxEvent, error) {
	var models []OutboxEventModel

	// SELECT * FROM outbox_events WHERE status = 'PENDING' FOR UPDATE SKIP LOCKED LIMIT ?
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("status = ?", string(domain.OutboxStatusPending)).
		Order("created_at ASC").
		Limit(limit).
		Find(&models).Error

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

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	// TODO: Implement GET
	return nil, nil
}

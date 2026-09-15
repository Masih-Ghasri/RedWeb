package postgres

import (
	"context"
	"time"

	"github.com/Masih-Ghasri/RedWeb/services/payment-service/internal/core/domain"
	"gorm.io/gorm"
)

type PaymentModel struct {
	ID        string `gorm:"primaryKey"`
	OrderID   string `gorm:"uniqueIndex"`
	Amount    float64
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (PaymentModel) TableName() string { return "payments" }

type PostgresPaymentRepository struct {
	db *gorm.DB
}

func NewPostgresPaymentRepository(db *gorm.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

func (r *PostgresPaymentRepository) Save(ctx context.Context, payment *domain.Payment) error {
	model := &PaymentModel{
		ID:        payment.ID,
		OrderID:   payment.OrderID,
		Amount:    payment.Amount,
		Status:    string(payment.Status),
		CreatedAt: payment.CreatedAt,
		UpdatedAt: payment.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(model).Error
}

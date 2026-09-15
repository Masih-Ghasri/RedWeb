package postgres

import (
	"context"
	"github.com/Masih-Ghasri/RedWeb/services/shipping-service/internal/core/domain"
	"gorm.io/gorm"
)

type PostgresShippingRepository struct {
	db *gorm.DB
}

func NewPostgresShippingRepository(db *gorm.DB) *PostgresShippingRepository {
	return &PostgresShippingRepository{db: db}
}

func (r *PostgresShippingRepository) Save(ctx context.Context, shipment *domain.Shipment) error {
	model := toShipmentModel(shipment)
	return r.db.WithContext(ctx).Create(model).Error
}

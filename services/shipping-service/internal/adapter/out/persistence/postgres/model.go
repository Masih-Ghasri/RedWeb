package postgres

import (
	"github.com/Masih-Ghasri/RedWeb/services/shipping-service/internal/core/domain"
	"time"
)

type ShipmentModel struct {
	ID           string `gorm:"primaryKey"`
	OrderID      string `gorm:"uniqueIndex"`
	TrackingCode string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (ShipmentModel) TableName() string { return "shipments" }

func toShipmentModel(s *domain.Shipment) *ShipmentModel {
	return &ShipmentModel{
		ID:           s.ID,
		OrderID:      s.OrderID,
		TrackingCode: s.TrackingCode,
		Status:       string(s.Status),
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}

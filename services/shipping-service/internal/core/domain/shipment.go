package domain

import (
	"time"
)

type ShipmentStatus string

const (
	ShipmentStatusPending    ShipmentStatus = "PENDING"
	ShipmentStatusProcessing ShipmentStatus = "PROCESSING"
	ShipmentStatusDispatched ShipmentStatus = "DISPATCHED"
)

type Shipment struct {
	ID           string
	OrderID      string
	TrackingCode string
	Status       ShipmentStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewShipment(id, orderID string) (*Shipment, error) {
	if orderID == "" {
		return nil, ErrInvalidOrderID
	}

	return &Shipment{
		ID:           id,
		OrderID:      orderID,
		TrackingCode: "",
		Status:       ShipmentStatusPending,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}, nil
}

func (s *Shipment) StartProcessing() {
	s.Status = ShipmentStatusProcessing
	s.UpdatedAt = time.Now().UTC()
}

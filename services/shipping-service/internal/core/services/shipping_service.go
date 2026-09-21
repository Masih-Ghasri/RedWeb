package services

import (
	"context"
	"fmt"
	"log"

	"github.com/Masih-Ghasri/RedWeb/services/shipping-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/shipping-service/internal/core/ports"
	"github.com/google/uuid"
)

type ShippingService struct {
	repo ports.ShippingRepository
}

func NewShippingService(repo ports.ShippingRepository) *ShippingService {
	return &ShippingService{
		repo: repo,
	}
}

func (s *ShippingService) ProcessShipment(ctx context.Context, cmd ports.ProcessShipmentCommand) error {
	shipmentID := uuid.New().String()

	shipment, err := domain.NewShipment(shipmentID, cmd.OrderID)
	if err != nil {
		return err
	}

	shipment.StartProcessing()

	shipment.TrackingCode = "TRK-" + shipment.ID[:8]

	if err := s.repo.Save(ctx, shipment); err != nil {
		return fmt.Errorf("failed to save shipment for order %s: %w", cmd.OrderID, err)
	}

	log.Printf("Shipment created successfully for Order: %s | Tracking: %s", cmd.OrderID, shipment.TrackingCode)
	return nil
}

package ports

import (
	"context"

	"github.com/Masih-Ghasri/RedWeb/services/shipping-service/internal/core/domain"
)

type ShippingRepository interface {
	Save(ctx context.Context, shipment *domain.Shipment) error
}

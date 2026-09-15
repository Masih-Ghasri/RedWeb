package ports

import (
	"context"

	"github.com/Masih-Ghasri/RedWeb/services/payment-service/internal/core/domain"
)

type PaymentRepository interface {
	Save(ctx context.Context, payment *domain.Payment) error
}

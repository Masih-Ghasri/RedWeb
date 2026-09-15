package ports

import (
	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/domain"

	"context"
)

type ProductRepository interface {
	Save(ctx context.Context, product *domain.Product) error
	FindByID(ctx context.Context, id string) (*domain.Product, error)
}

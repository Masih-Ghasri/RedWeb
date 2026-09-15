package ports

import (
	"context"

	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/domain"
)

type CreateProductCommand struct {
	Name        string
	Description string
	Price       float64
	Stock       int
}

type ProductUseCase interface {
	CreateProduct(ctx context.Context, cmd CreateProductCommand) (*domain.Product, error)
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
}

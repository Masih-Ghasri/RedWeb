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

// UpdateProductCommand carries the full new state of a product. Stock is
// allowed to be zero (marking something out of stock) but never negative —
// domain.Product.Update enforces that.
type UpdateProductCommand struct {
	Name        string
	Description string
	Price       float64
	Stock       int
}

// ListProductsQuery is the caller-facing view of domain.ProductFilter. The
// service normalises and clamps it before handing it to the repository.
type ListProductsQuery struct {
	Q           string
	MinPrice    *float64
	MaxPrice    *float64
	InStockOnly bool
	Page        int
	PageSize    int
}

type ProductList struct {
	Items    []*domain.Product
	Page     int
	PageSize int
	Total    int
}

type ProductUseCase interface {
	CreateProduct(ctx context.Context, cmd CreateProductCommand) (*domain.Product, error)
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
	ListProducts(ctx context.Context, q ListProductsQuery) (*ProductList, error)
	UpdateProduct(ctx context.Context, id string, cmd UpdateProductCommand) (*domain.Product, error)
	DeleteProduct(ctx context.Context, id string) error
}

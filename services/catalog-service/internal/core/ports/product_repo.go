package ports

import (
	"context"

	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/domain"
)

type ProductRepository interface {
	// Save inserts a new product. It does not upsert — updating goes through
	// Update so a missing document cannot be created by accident.
	Save(ctx context.Context, product *domain.Product) error

	// FindByID returns domain.ErrProductNotFound when the id does not exist.
	FindByID(ctx context.Context, id string) (*domain.Product, error)

	// FindAll returns the requested page together with the total number of
	// products matching the filter, which is what makes pagination possible
	// without a second round-trip in the caller.
	FindAll(ctx context.Context, filter domain.ProductFilter) ([]*domain.Product, int, error)

	// Update replaces an existing product and returns domain.ErrProductNotFound
	// if the id is unknown.
	Update(ctx context.Context, product *domain.Product) error

	// Delete removes a product and returns domain.ErrProductNotFound if the
	// id is unknown.
	Delete(ctx context.Context, id string) error
}

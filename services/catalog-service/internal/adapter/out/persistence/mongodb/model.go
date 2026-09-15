package mongodb

import (
	"time"

	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/domain"
)

// ProductDocument is the MongoDB representation of a product.
//
// Every field of domain.Product that matters to readers must be mapped here
// or it is silently dropped on the round-trip — Status used to be.
type ProductDocument struct {
	ID          string    `bson:"_id"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	Price       float64   `bson:"price"`
	Stock       int       `bson:"stock"`
	Status      string    `bson:"status"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

func toDocument(p *domain.Product) *ProductDocument {
	return &ProductDocument{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		Status:      p.Status,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toDomain(d *ProductDocument) *domain.Product {
	return &domain.Product{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		Price:       d.Price,
		Stock:       d.Stock,
		Status:      d.Status,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

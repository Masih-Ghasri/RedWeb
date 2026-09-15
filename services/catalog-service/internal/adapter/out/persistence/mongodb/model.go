package mongodb

import (
	"time"

	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/domain"
)

type ProductDocument struct {
	ID          string    `bson:"_id"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	Price       float64   `bson:"price"`
	Stock       int       `bson:"stock"`
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
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

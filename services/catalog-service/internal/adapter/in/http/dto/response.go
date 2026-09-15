package dto

import "github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/domain"

type ProductResponse struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

func FromDomain(p *domain.Product) ProductResponse {
	return ProductResponse{
		ID:    p.ID,
		Name:  p.Name,
		Price: p.Price,
		Stock: p.Stock,
	}
}

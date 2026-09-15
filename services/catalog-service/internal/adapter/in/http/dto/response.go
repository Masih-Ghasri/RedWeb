package dto

import (
	"time"

	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/ports"
)

// ProductResponse is the public shape of a product.
//
// Every domain field a client might render is present — Description and
// Status used to be dropped here, which made the API look like it had lost
// data that was in fact stored.
type ProductResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func FromDomain(p *domain.Product) ProductResponse {
	return ProductResponse{
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

// ProductListResponse is one page of a product listing. Total is the number
// of matches across all pages, not the length of Items.
type ProductListResponse struct {
	Items    []ProductResponse `json:"items"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Total    int               `json:"total"`
}

// FromProductList converts a whole page in one place so a handler cannot
// forget to surface the pagination metadata alongside the items.
func FromProductList(list *ports.ProductList) ProductListResponse {
	items := make([]ProductResponse, 0, len(list.Items))
	for _, p := range list.Items {
		items = append(items, FromDomain(p))
	}
	return ProductListResponse{
		Items:    items,
		Page:     list.Page,
		PageSize: list.PageSize,
		Total:    list.Total,
	}
}

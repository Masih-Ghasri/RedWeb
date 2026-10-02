package dto

import (
	"time"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/ports"
)

type OrderItemResponse struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// OrderResponse is the public response representation of an order.
type OrderResponse struct {
	ID          string              `json:"id"`
	CustomerID  string              `json:"customer_id"`
	Items       []OrderItemResponse `json:"items"`
	TotalAmount float64             `json:"total_amount"`
	Status      string              `json:"status"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

func FromDomain(o *domain.Order) OrderResponse {
	items := make([]OrderItemResponse, 0, len(o.Items))
	for _, item := range o.Items {
		items = append(items, OrderItemResponse{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	return OrderResponse{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Items:       items,
		TotalAmount: o.TotalAmount,
		Status:      string(o.Status),
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}

type OrderListResponse struct {
	Items    []OrderResponse `json:"items"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Total    int             `json:"total"`
}

// FromOrderList converts a whole page in one place so a handler cannot
// forget to surface the pagination metadata alongside the items.
func FromOrderList(list *ports.OrderList) OrderListResponse {
	items := make([]OrderResponse, 0, len(list.Items))
	for _, o := range list.Items {
		items = append(items, FromDomain(o))
	}

	return OrderListResponse{
		Items:    items,
		Page:     list.Page,
		PageSize: list.PageSize,
		Total:    list.Total,
	}
}

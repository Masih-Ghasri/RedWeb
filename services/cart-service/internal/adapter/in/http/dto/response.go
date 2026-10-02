package dto

import (
	"time"

	"github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/core/domain"
)

// CartItemResponse is one line of a cart as the API returns it. Subtotal is
// computed here because it is a presentation concern, not cart state.
type CartItemResponse struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Subtotal  float64 `json:"subtotal"`
}

type CartResponse struct {
	CustomerID string             `json:"customer_id"`
	Items      []CartItemResponse `json:"items"`
	Total      float64            `json:"total"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

// FromDomain converts a cart for the wire. An empty cart serialises as
// `"items": []` rather than `null`, so clients can iterate without a nil
// check.
func FromDomain(c *domain.Cart) CartResponse {
	items := make([]CartItemResponse, 0, len(c.Items))
	for _, item := range c.Items {
		items = append(items, CartItemResponse{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			Subtotal:  item.Price * float64(item.Quantity),
		})
	}

	return CartResponse{
		CustomerID: c.CustomerID,
		Items:      items,
		Total:      c.Total,
		UpdatedAt:  c.UpdatedAt,
	}
}

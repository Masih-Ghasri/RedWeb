package dto

// OrderItemRequest represents an individual line item in an order creation request.
type OrderItemRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"gt=0"`
}

// CreateOrderRequest represents the payload for creating an order.
type CreateOrderRequest struct {
	CartID     string             `json:"cart_id"`
	CustomerID string             `json:"customer_id"`
	Items      []OrderItemRequest `json:"items" validate:"omitempty,dive"`
}

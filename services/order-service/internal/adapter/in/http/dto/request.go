package dto

// request.go
type OrderItemRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
}

type CreateOrderRequest struct {
	CustomerID string             `json:"customer_id" validate:"required"`
	Items      []OrderItemRequest `json:"items" validate:"required,dive"`
}

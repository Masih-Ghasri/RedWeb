package dto

type OrderResponse struct {
	ID          string  `json:"id"`
	TotalAmount float64 `json:"total_amount"`
	Status      string  `json:"status"`
}

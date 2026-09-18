package domain

type OrderCreatedEvent struct {
	OrderID     string  `json:"order_id"`
	CustomerID  string  `json:"customer_id"`
	TotalAmount float64 `json:"total_amount"`
	Status      string  `json:"status"`
}

type PaymentProcessedEvent struct {
	PaymentID string `json:"payment_id"`
	OrderID   string `json:"order_id"`
	Status    string `json:"status"`
}

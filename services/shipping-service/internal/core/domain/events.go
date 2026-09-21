package domain

type PaymentProcessedEvent struct {
	PaymentID string `json:"payment_id"`
	OrderID   string `json:"order_id"`
	Status    string `json:"status"` // SUCCESS or FAILED
}

package domain

import "time"

type OrderCreatedEvent struct {
	OrderID     string  `json:"order_id"`
	CustomerID  string  `json:"customer_id"`
	TotalAmount float64 `json:"total_amount"`
	Status      string  `json:"status"`
}

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "PENDING"
	OutboxStatusPublished OutboxStatus = "PUBLISHED"
)

type OutboxEvent struct {
	ID            string
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       []byte
	Status        OutboxStatus
	CreatedAt     time.Time
}

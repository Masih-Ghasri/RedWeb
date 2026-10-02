package domain

import "time"

// Shared RabbitMQ exchange and routing keys.
const (
	OrderExchange = "order.exchange"

	RoutingKeyProductCreated = "product.created"
	RoutingKeyProductUpdated = "product.updated"
	RoutingKeyProductDeleted = "product.deleted"
)

type OrderCreatedEvent struct {
	OrderID     string  `json:"order_id"`
	CustomerID  string  `json:"customer_id"`
	TotalAmount float64 `json:"total_amount"`
	Status      string  `json:"status"`
}

type OutboxStatus string

const (
	OutboxStatusPending    OutboxStatus = "PENDING"
	OutboxStatusProcessing OutboxStatus = "PROCESSING"
	OutboxStatusPublished  OutboxStatus = "PUBLISHED"
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

// ProductEvent mirrors catalog-service's payload for product.created and
// ProductEvent represents product catalog mutation events consumed by the cache worker.
type ProductEvent struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Stock     int     `json:"stock"`
	Status    string  `json:"status"`
}

// ProductDeletedEvent mirrors catalog-service's payload for product.deleted.
type ProductDeletedEvent struct {
	ProductID string `json:"product_id"`
}

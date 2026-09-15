package ports

import "context"

// EventPublisher defines the contract for publishing domain events.
type EventPublisher interface {
	Publish(ctx context.Context, exchange, routingKey string, payload []byte) error
}

package rabbitmq

import (
	"context"
	"fmt"

	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQPublisher publishes catalog domain events onto the shared topic
// exchange consumed by order-service.
type RabbitMQPublisher struct {
	channel *amqp.Channel
}

// NewRabbitMQPublisher declares the shared exchange and returns a publisher
// bound to it.
//
// Declaration is idempotent in RabbitMQ: re-declaring an exchange with
// identical arguments is a no-op, so catalog-service starts correctly whether
// or not order-service has ever run. Returning an error (rather than logging
// a warning) is deliberate — without the exchange, every publish would fail
// silently and order-service's product cache would never warm up.
func NewRabbitMQPublisher(ch *amqp.Channel) (*RabbitMQPublisher, error) {
	if err := ch.ExchangeDeclare(
		domain.OrderExchange, // name
		"topic",              // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		nil,                  // arguments
	); err != nil {
		return nil, fmt.Errorf("declare exchange %q: %w", domain.OrderExchange, err)
	}

	return &RabbitMQPublisher{channel: ch}, nil
}

// Publish publishes a message with the specified routing key to the exchange.
func (p *RabbitMQPublisher) Publish(ctx context.Context, exchange, routingKey string, payload []byte) error {
	return p.channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         payload,
		},
	)
}

package rabbitmq

import (
	"context"
	"log"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	channel *amqp.Channel
}

// NewRabbitMQPublisher declares the shared exchange and returns a publisher
// bound to it.
//
// Declaration is idempotent, and failure is only a warning here because the
// product cache consumer re-declares the exchange before it binds: a broker
// that has never seen this exchange still ends up with it. Outbox publishing
// keeps working either way — only routing would be lost until a consumer
// declares it.
func NewRabbitMQPublisher(ch *amqp.Channel) *RabbitMQPublisher {
	if err := ch.ExchangeDeclare(
		domain.OrderExchange, // name
		"topic",              // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		nil,                  // arguments
	); err != nil {
		log.Printf("Warning: failed to declare exchange %q: %v", domain.OrderExchange, err)
	}

	return &RabbitMQPublisher{
		channel: ch,
	}
}

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

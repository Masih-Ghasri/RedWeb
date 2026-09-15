package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	channel *amqp.Channel
}

func NewRabbitMQPublisher(ch *amqp.Channel) *RabbitMQPublisher {
	err := ch.ExchangeDeclare(
		"order.exchange", // name
		"topic",          // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		fmt.Printf("Warning: failed to declare exchange: %v\n", err)
	}

	return &RabbitMQPublisher{
		channel: ch,
	}
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, exchange, routingKey string, payload []byte) error {
	return p.channel.PublishWithContext(ctx,
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

package rabbitmq

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	channel *amqp.Channel
}

func NewRabbitMQPublisher(ch *amqp.Channel) *RabbitMQPublisher {
	return &RabbitMQPublisher{channel: ch}
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, exchange, routingKey string, payload []byte) error {
	return p.channel.PublishWithContext(ctx,
		exchange, routingKey, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         payload,
		},
	)
}

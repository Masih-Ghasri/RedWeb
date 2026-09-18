package rabbitmq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Masih-Ghasri/RedWeb/services/payment-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/payment-service/internal/core/ports"
	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderCreatedConsumer struct {
	channel *amqp.Channel
	useCase ports.PaymentUseCase
}

func NewOrderCreatedConsumer(ch *amqp.Channel, uc ports.PaymentUseCase) *OrderCreatedConsumer {
	return &OrderCreatedConsumer{
		channel: ch,
		useCase: uc,
	}
}

func (c *OrderCreatedConsumer) Start(ctx context.Context) error {
	q, err := c.channel.QueueDeclare(
		"payment.order.created.queue", // name
		true,                          // durable
		false,                         // delete when unused
		false,                         // exclusive
		false,                         // no-wait
		nil,                           // arguments
	)
	if err != nil {
		return err
	}

	err = c.channel.QueueBind(
		q.Name,           // queue name
		"order.created",  // routing key
		"order.exchange", // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	msgs, err := c.channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	log.Println("Payment Consumer is listening for messages...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Shutting down consumer...")
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}
				c.handleMessage(ctx, d)
			}
		}
	}()

	return nil
}

func (c *OrderCreatedConsumer) handleMessage(ctx context.Context, d amqp.Delivery) {
	var event domain.OrderCreatedEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		log.Printf("Error decoding message: %v", err)
		d.Nack(false, false)
		return
	}

	cmd := ports.ProcessPaymentCommand{
		OrderID: event.OrderID,
		Amount:  event.TotalAmount,
	}

	err := c.useCase.ProcessPayment(ctx, cmd)
	if err != nil {
		log.Printf("Failed to process payment for order %s: %v", event.OrderID, err)
		d.Nack(false, true)
		return
	}

	d.Ack(false)
}

package rabbitmq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/ports"
	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentEventPayload struct {
	PaymentID string `json:"payment_id"`
	OrderID   string `json:"order_id"`
	Status    string `json:"status"` // SUCCESS or FAILED
}

type PaymentProcessedConsumer struct {
	channel *amqp.Channel
	useCase ports.OrderUseCase
}

func NewPaymentProcessedConsumer(ch *amqp.Channel, uc ports.OrderUseCase) *PaymentProcessedConsumer {
	return &PaymentProcessedConsumer{
		channel: ch,
		useCase: uc,
	}
}

func (c *PaymentProcessedConsumer) Start(ctx context.Context) error {
	q, err := c.channel.QueueDeclare(
		"order.payment.processed.queue",
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	err = c.channel.QueueBind(
		q.Name,
		"payment.processed",
		"order.exchange", // Exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	msgs, err := c.channel.Consume(
		q.Name, "", false, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	log.Println("Order Service: Listening for 'payment.processed' events...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Order Service Consumer shutting down...")
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

func (c *PaymentProcessedConsumer) handleMessage(ctx context.Context, d amqp.Delivery) {
	var payload PaymentEventPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		log.Printf("Invalid payment event payload: %v", err)
		d.Nack(false, false)
		return
	}

	newStatus := "PAID"
	if payload.Status == "FAILED" {
		newStatus = "FAILED"
	}

	err := c.useCase.UpdateOrderStatus(ctx, payload.OrderID, newStatus)
	if err != nil {
		log.Printf("Failed to update order %s: %v", payload.OrderID, err)
		d.Nack(false, false)
		return
	}

	log.Printf("Successfully updated order %s to status %s", payload.OrderID, newStatus)
	d.Ack(false)
}

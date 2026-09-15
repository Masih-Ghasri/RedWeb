package rabbitmq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Masih-Ghasri/RedWeb/services/shipping-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/shipping-service/internal/core/ports"
	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentProcessedConsumer struct {
	channel *amqp.Channel
	useCase ports.ShippingUseCase
}

func NewPaymentProcessedConsumer(ch *amqp.Channel, uc ports.ShippingUseCase) *PaymentProcessedConsumer {
	return &PaymentProcessedConsumer{
		channel: ch,
		useCase: uc,
	}
}

func (c *PaymentProcessedConsumer) Start(ctx context.Context) error {
	q, err := c.channel.QueueDeclare(
		"shipping.payment.processed.queue",
		true, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	err = c.channel.QueueBind(
		q.Name,
		"payment.processed",
		"order.exchange",
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

	log.Println("Shipping Consumer is listening for successful payments...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Shipping Consumer shutting down...")
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
	var event domain.PaymentProcessedEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		log.Printf("Invalid event payload: %v", err)
		d.Nack(false, false)
		return
	}

	if event.Status != "SUCCESS" {
		log.Printf("Ignoring payment event for order %s because status is %s", event.OrderID, event.Status)
		d.Ack(false)
		return
	}

	cmd := ports.ProcessShipmentCommand{
		OrderID: event.OrderID,
	}

	err := c.useCase.ProcessShipment(ctx, cmd)
	if err != nil {
		log.Printf("Failed to process shipment for order %s: %v", event.OrderID, err)
		d.Nack(false, true)
		return
	}

	d.Ack(false)
}

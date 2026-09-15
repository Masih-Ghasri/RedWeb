package services

import (
	"context"
	"log"
	"time"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/ports"
)

type OutboxRelayWorker struct {
	outboxRepo ports.OutboxRepository
	publisher  ports.EventPublisher
	ticker     *time.Ticker
}

func NewOutboxRelayWorker(repo ports.OutboxRepository, pub ports.EventPublisher, interval time.Duration) *OutboxRelayWorker {
	return &OutboxRelayWorker{
		outboxRepo: repo,
		publisher:  pub,
		ticker:     time.NewTicker(interval),
	}
}

func (w *OutboxRelayWorker) Start(ctx context.Context) {
	log.Println("Outbox Relay Worker started...")
	for {
		select {
		case <-ctx.Done():
			log.Println("Outbox Relay Worker stopping due to context cancellation...")
			w.ticker.Stop()
			return
		case <-w.ticker.C:
			w.processOutbox(ctx)
		}
	}
}

func (w *OutboxRelayWorker) processOutbox(ctx context.Context) {
	events, err := w.outboxRepo.GetPendingEvents(ctx, 50)
	if err != nil {
		log.Printf("Error fetching pending outbox events: %v", err)
		return
	}

	for _, event := range events {
		routingKey := "order.created"
		if event.EventType != "OrderCreated" {
			routingKey = "order.misc"
		}

		err := w.publisher.Publish(ctx, "order.exchange", routingKey, event.Payload)
		if err != nil {
			log.Printf("Failed to publish event %s: %v", event.ID, err)
			continue
		}

		if err := w.outboxRepo.MarkAsPublished(ctx, event.ID); err != nil {
			log.Printf("Failed to mark event %s as published: %v", event.ID, err)
		} else {
			log.Printf("Successfully published event %s to RabbitMQ", event.ID)
		}
	}
}

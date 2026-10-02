package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/ports"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// productCacheQueue receives every product mutation catalog-service
	// publishes. One queue for all three routing keys keeps ordering
	// per-product: a create, an update and a delete are always applied in the
	// order they were published.
	productCacheQueue = "order.product.cache.queue"
)

// ProductConsumer turns catalog-service product events into entries in the
// shared Redis product cache that CreateOrder prices orders from.
type ProductConsumer struct {
	channel *amqp.Channel
	cache   ports.ProductCache
}

func NewProductConsumer(ch *amqp.Channel, cache ports.ProductCache) *ProductConsumer {
	return &ProductConsumer{
		channel: ch,
		cache:   cache,
	}
}

// Start declares the topology, begins consuming and returns immediately; the
// delivery loop runs on its own goroutine and stops when ctx is cancelled.
func (c *ProductConsumer) Start(ctx context.Context) error {
	// Declaring the exchange here (idempotently) means this consumer works on
	// a broker that catalog-service has never touched, instead of failing the
	// bind with a 404 channel exception.
	if err := c.channel.ExchangeDeclare(
		domain.OrderExchange, // name
		"topic",              // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		nil,                  // arguments
	); err != nil {
		return fmt.Errorf("declare exchange %q: %w", domain.OrderExchange, err)
	}

	q, err := c.channel.QueueDeclare(
		productCacheQueue, // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return fmt.Errorf("declare queue %q: %w", productCacheQueue, err)
	}

	for _, routingKey := range []string{
		domain.RoutingKeyProductCreated,
		domain.RoutingKeyProductUpdated,
		domain.RoutingKeyProductDeleted,
	} {
		if err := c.channel.QueueBind(q.Name, routingKey, domain.OrderExchange, false, nil); err != nil {
			return fmt.Errorf("bind %q to queue %q: %w", routingKey, q.Name, err)
		}
	}

	msgs, err := c.channel.Consume(
		q.Name, // queue
		"",     // consumer tag
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("start consumer on queue %q: %w", q.Name, err)
	}

	log.Printf("Order Service: product cache consumer listening on %q", productCacheQueue)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Product cache consumer shutting down...")
				return
			case d, ok := <-msgs:
				if !ok {
					log.Println("Product cache consumer stopped: delivery channel closed")
					return
				}
				c.handleMessage(ctx, d)
			}
		}
	}()

	return nil
}

func (c *ProductConsumer) handleMessage(ctx context.Context, d amqp.Delivery) {
	switch d.RoutingKey {
	case domain.RoutingKeyProductCreated, domain.RoutingKeyProductUpdated:
		var event domain.ProductEvent
		if err := json.Unmarshal(d.Body, &event); err != nil {
			log.Printf("Product cache: invalid %s payload: %v", d.RoutingKey, err)
			nackDelivery(d, false)
			return
		}
		if event.ProductID == "" {
			log.Printf("Product cache: %s payload carries no product_id, discarding", d.RoutingKey)
			nackDelivery(d, false)
			return
		}

		if err := c.cache.SetProduct(ctx, event.ProductID, event.Stock, event.Price); err != nil {
			log.Printf("Product cache: failed to cache product %s: %v", event.ProductID, err)
			nackDelivery(d, true)
			return
		}

		ackDelivery(d)

	case domain.RoutingKeyProductDeleted:
		var event domain.ProductDeletedEvent
		if err := json.Unmarshal(d.Body, &event); err != nil {
			log.Printf("Product cache: invalid %s payload: %v", d.RoutingKey, err)
			nackDelivery(d, false)
			return
		}
		if event.ProductID == "" {
			log.Printf("Product cache: %s payload carries no product_id, discarding", d.RoutingKey)
			nackDelivery(d, false)
			return
		}

		if err := c.cache.DeleteProduct(ctx, event.ProductID); err != nil {
			log.Printf("Product cache: failed to evict product %s: %v", event.ProductID, err)
			nackDelivery(d, true)
			return
		}

		ackDelivery(d)

	default:
		// Unreachable while the binds above are the only way onto this queue.
		log.Printf("Product cache: unexpected routing key %q, discarding", d.RoutingKey)
		nackDelivery(d, false)
	}
}

// ackDelivery and nackDelivery record ack failures instead of dropping them
// silently. A lost ack makes the broker redeliver, and an invisible redelivery
// is the hardest kind of duplicate to chase down later.
func ackDelivery(d amqp.Delivery) {
	if err := d.Ack(false); err != nil {
		log.Printf("Product cache: failed to acknowledge message: %v", err)
	}
}

func nackDelivery(d amqp.Delivery, requeue bool) {
	if err := d.Nack(false, requeue); err != nil {
		log.Printf("Product cache: failed to negative-acknowledge message (requeue=%t): %v", requeue, err)
	}
}

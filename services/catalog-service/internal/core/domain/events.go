package domain

// Routing keys published by catalog-service onto the shared "order.exchange"
// (topic) exchange. order-service binds its product-cache consumer to these.
const (
	RoutingKeyProductCreated = "product.created"
	RoutingKeyProductUpdated = "product.updated"
	RoutingKeyProductDeleted = "product.deleted"
)

// OrderExchange is the topic exchange shared by every service in RedWeb.
// It is declared idempotently by each service that publishes to, or binds
// against, it — so no service depends on another having started first.
const OrderExchange = "order.exchange"

// ProductEvent is the payload for product.created and product.updated.
//
// Price and Stock are the two fields order-service writes into its Redis
// product cache (key "product:{product_id}"), which CreateOrder reads to
// price and validate an order. Keep those field names stable: changing them
// silently breaks checkout.
type ProductEvent struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Stock     int     `json:"stock"`
	Status    string  `json:"status"`
}

// ProductDeletedEvent is the payload for product.deleted. The consumer uses
// it to evict the entry from the product cache.
type ProductDeletedEvent struct {
	ProductID string `json:"product_id"`
}

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/ports"
)

// clearCartTimeout bounds the post-checkout cart deletion. It is independent
// of any request context on purpose — see OrderService.clearCart.
const clearCartTimeout = 5 * time.Second

type OrderService struct {
	repo  ports.OrderRepository
	cache ports.InventoryCache
	cart  ports.CartClient
}

func NewOrderService(repo ports.OrderRepository, cache ports.InventoryCache, cart ports.CartClient) *OrderService {
	return &OrderService{
		repo:  repo,
		cache: cache,
		cart:  cart,
	}
}

// CreateOrder turns either a cart or an explicit item list into a persisted
// order plus its outbox event.
//
// The cart path is the one clients are meant to use: identity and line items
// both come from the cart, so neither is attacker-controlled.
func (s *OrderService) CreateOrder(ctx context.Context, cmd ports.CreateOrderCommand) (*domain.Order, error) {
	refs, customerID, err := s.resolveLines(ctx, cmd)
	if err != nil {
		return nil, err
	}

	items, err := s.pricedItems(ctx, refs)
	if err != nil {
		return nil, err
	}

	orderID := uuid.New().String()
	order, err := domain.NewOrder(orderID, customerID, items)
	if err != nil {
		return nil, err
	}

	eventPayload := domain.OrderCreatedEvent{
		OrderID:     order.ID,
		CustomerID:  order.CustomerID,
		TotalAmount: order.TotalAmount,
		Status:      string(order.Status),
	}

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return nil, fmt.Errorf("encode order event: %w", err)
	}

	outboxEvent := &domain.OutboxEvent{
		ID:            uuid.New().String(),
		AggregateType: "ORDER",
		AggregateID:   order.ID,
		EventType:     "OrderCreated",
		Payload:       payloadBytes,
		Status:        domain.OutboxStatusPending,
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.repo.CreateOrderWithOutbox(ctx, order, outboxEvent); err != nil {
		return nil, fmt.Errorf("failed to process order transaction: %w", err)
	}

	// Only after the transaction has committed, and never fatal: the order is
	// durable at this point, and failing a completed checkout because a cart
	// could not be cleared would be worse than a cart that lingers until its
	// TTL expires.
	if cmd.CartID != "" {
		s.clearCart(cmd.CartID)
	}

	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	return s.repo.GetByID(ctx, id)
}

// ListOrders clamps pagination and returns one page plus the matching total.
func (s *OrderService) ListOrders(ctx context.Context, q ports.ListOrdersQuery) (*ports.OrderList, error) {
	page := q.Page
	if page < 1 {
		page = 1
	}

	pageSize := q.PageSize
	if pageSize < 1 {
		pageSize = domain.DefaultPageSize
	}
	if pageSize > domain.MaxPageSize {
		pageSize = domain.MaxPageSize
	}

	orders, total, err := s.repo.FindAll(ctx, ports.OrderFilter{
		CustomerID: q.CustomerID,
		Status:     q.Status,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		return nil, err
	}

	return &ports.OrderList{
		Items:    orders,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

// CancelOrder transitions a pending order to CANCELED status.
func (s *OrderService) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err // domain.ErrOrderNotFound → 404
	}

	if order.Status != domain.StatusPending {
		return nil, domain.ErrOrderNotCancelable
	}

	changed, err := s.repo.CancelIfPending(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("cancel order %s: %w", id, err)
	}
	if !changed {
		return nil, domain.ErrOrderNotCancelable
	}

	order.Status = domain.StatusCanceled
	order.UpdatedAt = time.Now().UTC()

	return order, nil
}

// UpdateOrderStatus is the payment consumer's entry point.
//
// A status update that matches no row is reported as an error rather than
// quietly succeeding: the order simply is not there, and the old code Ack'd
// straight past that, losing the transition with no trace in the logs.
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
	orderStatus := domain.OrderStatus(status)

	if orderStatus != domain.StatusPaid && orderStatus != domain.StatusCanceled && orderStatus != domain.StatusFailed {
		return fmt.Errorf("invalid status update requested: %s", status)
	}

	changed, err := s.repo.UpdateStatus(ctx, orderID, orderStatus)
	if err != nil {
		return fmt.Errorf("failed to update order %s to %s: %w", orderID, status, err)
	}
	if !changed {
		return fmt.Errorf("%w: cannot set order %s to %s", domain.ErrOrderNotFound, orderID, status)
	}

	return nil
}

// resolveLines decides where an order's shape comes from.
//
// CartID wins whenever it is present: the snapshot supplies both the line
// items and the customer id, which is what stops a caller from naming someone
// else as the customer. The legacy branch still honours cmd.CustomerID for
// pre-Phase-4 callers and is scheduled for removal.
func (s *OrderService) resolveLines(ctx context.Context, cmd ports.CreateOrderCommand) ([]ports.CartItemRef, string, error) {
	if cmd.CartID == "" {
		refs := make([]ports.CartItemRef, 0, len(cmd.Items))
		for _, item := range cmd.Items {
			refs = append(refs, ports.CartItemRef{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			})
		}
		return refs, cmd.CustomerID, nil
	}

	snapshot, err := s.cart.GetCart(ctx, cmd.CartID)
	if err != nil {
		return nil, "", err // ErrCartNotFound → 404, ErrCartUnavailable → 503
	}

	return snapshot.Items, snapshot.CustomerID, nil
}

// pricedItems replaces each product id with the live price from the product
// cache and refuses quantities beyond what is stocked.
//
// Prices come from the cache rather than the cart on purpose: a cart can sit
// for days while the catalogue moves, and charging an order at a stale price
// would be the wrong kind of surprise.
func (s *OrderService) pricedItems(ctx context.Context, refs []ports.CartItemRef) ([]domain.OrderItem, error) {
	items := make([]domain.OrderItem, 0, len(refs))

	for _, ref := range refs {
		stock, price, err := s.cache.GetProductStockAndPrice(ctx, ref.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product %s check failed: %w", ref.ProductID, err)
		}

		if stock < ref.Quantity {
			return nil, domain.ErrInsufficientStock
		}

		items = append(items, domain.OrderItem{
			ProductID: ref.ProductID,
			Quantity:  ref.Quantity,
			Price:     price,
		})
	}

	return items, nil
}

// clearCart empties the cart after a successful checkout.
//
// It runs on its own context so a client that disconnects the instant it
// receives `201` cannot cancel the cleanup, and it reports failure only to
// the log — the order is already committed.
func (s *OrderService) clearCart(cartID string) {
	ctx, cancel := context.WithTimeout(context.Background(), clearCartTimeout)
	defer cancel()

	if err := s.cart.ClearCart(ctx, cartID); err != nil {
		log.Printf("order: could not clear cart %s after checkout: %v", cartID, err)
	}
}

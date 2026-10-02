package services_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/ports"
	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/services"
)

// fakeOrderRepo is an in-memory implementation of ports.OrderRepository.
type fakeOrderRepo struct {
	mu     sync.Mutex
	orders map[string]*domain.Order
	events []*domain.OutboxEvent
}

func newFakeOrderRepo() *fakeOrderRepo {
	return &fakeOrderRepo{
		orders: make(map[string]*domain.Order),
	}
}

func (r *fakeOrderRepo) CreateOrderWithOutbox(ctx context.Context, order *domain.Order, event *domain.OutboxEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	copied := *order
	copied.Items = append([]domain.OrderItem(nil), order.Items...)
	r.orders[order.ID] = &copied
	r.events = append(r.events, event)
	return nil
}

func (r *fakeOrderRepo) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	o, ok := r.orders[id]
	if !ok {
		return nil, domain.ErrOrderNotFound
	}
	copied := *o
	copied.Items = append([]domain.OrderItem(nil), o.Items...)
	return &copied, nil
}

func (r *fakeOrderRepo) FindAll(ctx context.Context, f ports.OrderFilter) ([]*domain.Order, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var matched []*domain.Order
	for _, o := range r.orders {
		if f.CustomerID != "" && o.CustomerID != f.CustomerID {
			continue
		}
		if f.Status != "" && string(o.Status) != f.Status {
			continue
		}
		matched = append(matched, o)
	}
	return matched, len(matched), nil
}

func (r *fakeOrderRepo) CancelIfPending(ctx context.Context, id string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	o, ok := r.orders[id]
	if !ok {
		return false, nil
	}
	if o.Status != domain.StatusPending {
		return false, nil
	}
	o.Status = domain.StatusCanceled
	return true, nil
}

func (r *fakeOrderRepo) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	o, ok := r.orders[id]
	if !ok {
		return false, nil
	}
	o.Status = status
	return true, nil
}

// fakeInventoryCache is an in-memory stub for ports.InventoryCache.
type fakeInventoryCache struct {
	mu       sync.Mutex
	products map[string]struct {
		stock int
		price float64
	}
}

func newFakeInventoryCache() *fakeInventoryCache {
	return &fakeInventoryCache{
		products: make(map[string]struct {
			stock int
			price float64
		}),
	}
}

func (c *fakeInventoryCache) set(id string, stock int, price float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.products[id] = struct {
		stock int
		price float64
	}{stock: stock, price: price}
}

func (c *fakeInventoryCache) GetProductStockAndPrice(ctx context.Context, productID string) (int, float64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	p, ok := c.products[productID]
	if !ok {
		return 0, 0, domain.ErrProductNotFound
	}
	return p.stock, p.price, nil
}

// fakeCartClient is an in-memory stub for ports.CartClient.
type fakeCartClient struct {
	mu      sync.Mutex
	carts   map[string]*ports.CartSnapshot
	cleared []string
}

func newFakeCartClient() *fakeCartClient {
	return &fakeCartClient{
		carts: make(map[string]*ports.CartSnapshot),
	}
}

func (c *fakeCartClient) set(id, customerID string, items []ports.CartItemRef) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.carts[id] = &ports.CartSnapshot{
		CustomerID: customerID,
		Items:      items,
	}
}

func (c *fakeCartClient) GetCart(ctx context.Context, cartID string) (*ports.CartSnapshot, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s, ok := c.carts[cartID]
	if !ok {
		return nil, ports.ErrCartNotFound
	}
	return s, nil
}

func (c *fakeCartClient) ClearCart(ctx context.Context, cartID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cleared = append(c.cleared, cartID)
	delete(c.carts, cartID)
	return nil
}

func TestOrderService_CreateOrder_FromCart(t *testing.T) {
	ctx := context.Background()
	repo := newFakeOrderRepo()
	cache := newFakeInventoryCache()
	cartClient := newFakeCartClient()
	svc := services.NewOrderService(repo, cache, cartClient)

	// Seed cache
	cache.set("p1", 10, 25.0)
	cache.set("p2", 5, 40.0)

	// Seed cart
	cartClient.set("cart-alice", "alice", []ports.CartItemRef{
		{ProductID: "p1", Quantity: 2},
		{ProductID: "p2", Quantity: 1},
	})

	// 1. Create order from cart
	order, err := svc.CreateOrder(ctx, ports.CreateOrderCommand{
		CartID: "cart-alice",
	})
	if err != nil {
		t.Fatalf("failed to create order: %v", err)
	}
	if order.CustomerID != "alice" {
		t.Errorf("expected customer alice, got %s", order.CustomerID)
	}
	expectedTotal := (2 * 25.0) + (1 * 40.0) // 90.0
	if order.TotalAmount != expectedTotal {
		t.Errorf("expected total %f, got %f", expectedTotal, order.TotalAmount)
	}
	if order.Status != domain.StatusPending {
		t.Errorf("expected status PENDING, got %s", order.Status)
	}

	// 2. Cancel order
	canceled, err := svc.CancelOrder(ctx, order.ID)
	if err != nil {
		t.Fatalf("failed to cancel order: %v", err)
	}
	if canceled.Status != domain.StatusCanceled {
		t.Errorf("expected status CANCELED, got %s", canceled.Status)
	}

	// 3. Second cancel should fail with ErrOrderNotCancelable (409)
	_, err = svc.CancelOrder(ctx, order.ID)
	if !errors.Is(err, domain.ErrOrderNotCancelable) {
		t.Errorf("expected ErrOrderNotCancelable, got %v", err)
	}
}

func TestOrderService_CreateOrder_InsufficientStock(t *testing.T) {
	ctx := context.Background()
	repo := newFakeOrderRepo()
	cache := newFakeInventoryCache()
	cartClient := newFakeCartClient()
	svc := services.NewOrderService(repo, cache, cartClient)

	cache.set("p1", 2, 25.0) // only 2 in stock

	cartClient.set("cart-bob", "bob", []ports.CartItemRef{
		{ProductID: "p1", Quantity: 5}, // requests 5
	})

	_, err := svc.CreateOrder(ctx, ports.CreateOrderCommand{
		CartID: "cart-bob",
	})
	if !errors.Is(err, domain.ErrInsufficientStock) {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}
}

func TestOrderService_GetOrder_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeOrderRepo()
	cache := newFakeInventoryCache()
	cartClient := newFakeCartClient()
	svc := services.NewOrderService(repo, cache, cartClient)

	_, err := svc.GetOrder(ctx, "nonexistent-id")
	if !errors.Is(err, domain.ErrOrderNotFound) {
		t.Fatalf("expected ErrOrderNotFound, got %v", err)
	}
}

package services_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/core/ports"
	"github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/core/services"
)

// fakeCartRepo is an in-memory implementation of ports.CartRepository.
type fakeCartRepo struct {
	mu    sync.Mutex
	carts map[string]*domain.Cart
}

func newFakeCartRepo() *fakeCartRepo {
	return &fakeCartRepo{
		carts: make(map[string]*domain.Cart),
	}
}

func (r *fakeCartRepo) Get(ctx context.Context, customerID string) (*domain.Cart, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.carts[customerID]
	if !ok {
		return nil, nil // absence is nil, nil
	}
	// return copy
	copied := *c
	copied.Items = append([]domain.CartItem(nil), c.Items...)
	return &copied, nil
}

func (r *fakeCartRepo) Save(ctx context.Context, cart *domain.Cart) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	copied := *cart
	copied.Items = append([]domain.CartItem(nil), cart.Items...)
	r.carts[cart.CustomerID] = &copied
	return nil
}

func (r *fakeCartRepo) Delete(ctx context.Context, customerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.carts, customerID)
	return nil
}

// fakeProductReader simulates reading product info from the cache.
type fakeProductReader struct {
	mu       sync.Mutex
	products map[string]struct {
		stock int
		price float64
	}
}

func newFakeProductReader() *fakeProductReader {
	return &fakeProductReader{
		products: make(map[string]struct {
			stock int
			price float64
		}),
	}
}

func (r *fakeProductReader) set(productID string, stock int, price float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.products[productID] = struct {
		stock int
		price float64
	}{stock: stock, price: price}
}

func (r *fakeProductReader) GetProductInfo(ctx context.Context, productID string) (int, float64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.products[productID]
	if !ok {
		return 0, 0, domain.ErrProductNotFound
	}
	return p.stock, p.price, nil
}

func TestCartService_Lifecycle(t *testing.T) {
	ctx := context.Background()
	repo := newFakeCartRepo()
	reader := newFakeProductReader()
	svc := services.NewCartService(repo, reader)

	// Seed reader products
	reader.set("prod-1", 10, 20.0)
	reader.set("prod-2", 5, 50.0)

	customerID := "cust-1"

	// 1. Get empty cart
	cart, err := svc.GetCart(ctx, customerID)
	if err != nil {
		t.Fatalf("failed to get empty cart: %v", err)
	}
	if !cart.IsEmpty() {
		t.Errorf("expected empty cart")
	}

	// 2. Add item within stock
	cart, err = svc.AddItem(ctx, ports.AddItemCommand{
		CustomerID: customerID,
		ProductID:  "prod-1",
		Quantity:   3,
	})
	if err != nil {
		t.Fatalf("failed to add item: %v", err)
	}
	if cart.QuantityOf("prod-1") != 3 {
		t.Errorf("expected 3 items of prod-1, got %d", cart.QuantityOf("prod-1"))
	}
	if cart.Total != 60.0 {
		t.Errorf("expected total 60.0, got %f", cart.Total)
	}

	// 3. Add item exceeding stock
	_, err = svc.AddItem(ctx, ports.AddItemCommand{
		CustomerID: customerID,
		ProductID:  "prod-1",
		Quantity:   8, // 3 + 8 = 11 > 10 stock
	})
	if !errors.Is(err, domain.ErrInsufficientStock) {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}

	// 4. Update quantity
	cart, err = svc.UpdateItemQuantity(ctx, ports.UpdateItemCommand{
		CustomerID: customerID,
		ProductID:  "prod-1",
		Quantity:   5,
	})
	if err != nil {
		t.Fatalf("failed to update quantity: %v", err)
	}
	if cart.QuantityOf("prod-1") != 5 {
		t.Errorf("expected 5 items, got %d", cart.QuantityOf("prod-1"))
	}

	// 5. Update quantity to 0 removes line
	cart, err = svc.UpdateItemQuantity(ctx, ports.UpdateItemCommand{
		CustomerID: customerID,
		ProductID:  "prod-1",
		Quantity:   0,
	})
	if err != nil {
		t.Fatalf("failed to remove item via 0 qty: %v", err)
	}
	if !cart.IsEmpty() {
		t.Errorf("cart should be empty after setting qty 0")
	}

	// 6. Add non-existent product returns ErrProductNotFound
	_, err = svc.AddItem(ctx, ports.AddItemCommand{
		CustomerID: customerID,
		ProductID:  "unknown-prod",
		Quantity:   1,
	})
	if !errors.Is(err, domain.ErrProductNotFound) {
		t.Errorf("expected ErrProductNotFound, got %v", err)
	}

	// 7. ClearCart deletes cart
	_ = cart.AddItem("prod-2", 1, 50.0)
	_ = repo.Save(ctx, cart)
	err = svc.ClearCart(ctx, customerID)
	if err != nil {
		t.Fatalf("failed to clear cart: %v", err)
	}
	cart, _ = svc.GetCart(ctx, customerID)
	if !cart.IsEmpty() {
		t.Errorf("cart should be empty after clear")
	}
}

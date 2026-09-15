package services_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/ports"
	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/services"
)

// fakeProductRepo is an in-memory implementation of ports.ProductRepository.
type fakeProductRepo struct {
	mu       sync.Mutex
	products map[string]*domain.Product
}

func newFakeProductRepo() *fakeProductRepo {
	return &fakeProductRepo{
		products: make(map[string]*domain.Product),
	}
}

func (r *fakeProductRepo) Save(ctx context.Context, product *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.products[product.ID] = product
	return nil
}

func (r *fakeProductRepo) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.products[id]
	if !ok {
		return nil, domain.ErrProductNotFound
	}
	return p, nil
}

func (r *fakeProductRepo) FindAll(ctx context.Context, f domain.ProductFilter) ([]*domain.Product, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var matched []*domain.Product
	for _, p := range r.products {
		if f.Q != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(f.Q)) {
			continue
		}
		if f.MinPrice != nil && p.Price < *f.MinPrice {
			continue
		}
		if f.MaxPrice != nil && p.Price > *f.MaxPrice {
			continue
		}
		if f.InStockOnly && p.Stock <= 0 {
			continue
		}
		matched = append(matched, p)
	}

	total := len(matched)
	// simple pagination for in-memory test
	start := (f.Page - 1) * f.PageSize
	if start >= total {
		return []*domain.Product{}, total, nil
	}
	end := start + f.PageSize
	if end > total {
		end = total
	}

	return matched[start:end], total, nil
}

func (r *fakeProductRepo) Update(ctx context.Context, product *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.products[product.ID]; !ok {
		return domain.ErrProductNotFound
	}
	r.products[product.ID] = product
	return nil
}

func (r *fakeProductRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.products[id]; !ok {
		return domain.ErrProductNotFound
	}
	delete(r.products, id)
	return nil
}

// fakeEventPublisher captures published events for test assertions.
type fakeEventPublisher struct {
	mu     sync.Mutex
	events []publishedEvent
}

type publishedEvent struct {
	Exchange   string
	RoutingKey string
	Payload    []byte
}

func newFakeEventPublisher() *fakeEventPublisher {
	return &fakeEventPublisher{}
}

func (p *fakeEventPublisher) Publish(ctx context.Context, exchange, routingKey string, payload []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, publishedEvent{
		Exchange:   exchange,
		RoutingKey: routingKey,
		Payload:    payload,
	})
	return nil
}

func (p *fakeEventPublisher) lastRoutingKey() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.events) == 0 {
		return ""
	}
	return p.events[len(p.events)-1].RoutingKey
}

func TestProductService_Lifecycle(t *testing.T) {
	ctx := context.Background()
	repo := newFakeProductRepo()
	pub := newFakeEventPublisher()
	svc := services.NewProductService(repo, pub)

	// 1. Create product
	cmd := ports.CreateProductCommand{
		Name:        "Wireless Mouse",
		Description: "Ergonomic 2.4G",
		Price:       29.99,
		Stock:       15,
	}
	created, err := svc.CreateProduct(ctx, cmd)
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}
	if created.ID == "" {
		t.Errorf("product ID must not be empty")
	}
	if pub.lastRoutingKey() != domain.RoutingKeyProductCreated {
		t.Errorf("expected routing key %s, got %s", domain.RoutingKeyProductCreated, pub.lastRoutingKey())
	}

	// 2. Get product
	found, err := svc.GetProduct(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get product: %v", err)
	}
	if found.Name != cmd.Name {
		t.Errorf("expected name %s, got %s", cmd.Name, found.Name)
	}

	// 3. List products
	list, err := svc.ListProducts(ctx, ports.ListProductsQuery{
		Q:        "Mouse",
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("failed to list products: %v", err)
	}
	if list.Total != 1 || len(list.Items) != 1 {
		t.Errorf("expected 1 item, got total %d, items %d", list.Total, len(list.Items))
	}

	// 4. Update product
	updated, err := svc.UpdateProduct(ctx, created.ID, ports.UpdateProductCommand{
		Name:        "Wireless Mouse Pro",
		Description: "Ergonomic Bluetooth + 2.4G",
		Price:       39.99,
		Stock:       10,
	})
	if err != nil {
		t.Fatalf("failed to update product: %v", err)
	}
	if updated.Name != "Wireless Mouse Pro" {
		t.Errorf("expected updated name, got %s", updated.Name)
	}
	if pub.lastRoutingKey() != domain.RoutingKeyProductUpdated {
		t.Errorf("expected routing key %s, got %s", domain.RoutingKeyProductUpdated, pub.lastRoutingKey())
	}

	// 5. Delete product
	err = svc.DeleteProduct(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to delete product: %v", err)
	}
	if pub.lastRoutingKey() != domain.RoutingKeyProductDeleted {
		t.Errorf("expected routing key %s, got %s", domain.RoutingKeyProductDeleted, pub.lastRoutingKey())
	}

	// 6. Verify deleted product cannot be found
	_, err = svc.GetProduct(ctx, created.ID)
	if err != domain.ErrProductNotFound {
		t.Errorf("expected ErrProductNotFound, got %v", err)
	}
}

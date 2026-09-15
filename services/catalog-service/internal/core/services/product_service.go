package services

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/ports"
	"github.com/google/uuid"
)

// Pagination defaults are defined once in the domain package
// (domain.DefaultPageSize / domain.MaxPageSize) so the service and the
// persistence adapter can never drift apart.

type ProductService struct {
	repo      ports.ProductRepository
	publisher ports.EventPublisher
}

func NewProductService(repo ports.ProductRepository, publisher ports.EventPublisher) *ProductService {
	return &ProductService{
		repo:      repo,
		publisher: publisher,
	}
}

// CreateProduct validates the incoming command, persists the product and
// announces it on the shared exchange so order-service can warm its product
// cache.
func (s *ProductService) CreateProduct(ctx context.Context, cmd ports.CreateProductCommand) (*domain.Product, error) {
	id := uuid.New().String()

	product, err := domain.NewProduct(id, cmd.Name, cmd.Description, cmd.Price, cmd.Stock)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, product); err != nil {
		return nil, err
	}

	s.announce(ctx, domain.RoutingKeyProductCreated, productEvent(product))

	return product, nil
}

func (s *ProductService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	return s.repo.FindByID(ctx, id)
}

// ListProducts clamps the caller's pagination, applies the filter and returns
// one page plus the matching total.
func (s *ProductService) ListProducts(ctx context.Context, q ports.ListProductsQuery) (*ports.ProductList, error) {
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

	items, total, err := s.repo.FindAll(ctx, domain.ProductFilter{
		Q:           q.Q,
		MinPrice:    q.MinPrice,
		MaxPrice:    q.MaxPrice,
		InStockOnly: q.InStockOnly,
		Page:        page,
		PageSize:    pageSize,
	})
	if err != nil {
		return nil, err
	}

	return &ports.ProductList{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

// UpdateProduct re-reads the product, applies the change and broadcasts the
// new state. Re-reading first means validation always runs against current
// values and a stale client cannot resurrect an already-deleted product.
func (s *ProductService) UpdateProduct(ctx context.Context, id string, cmd ports.UpdateProductCommand) (*domain.Product, error) {
	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := product.Update(cmd.Name, cmd.Description, cmd.Price, cmd.Stock); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, product); err != nil {
		return nil, err
	}

	s.announce(ctx, domain.RoutingKeyProductUpdated, productEvent(product))

	return product, nil
}

// DeleteProduct removes the product and tells order-service to evict it from
// its cache, so a deleted product can no longer be priced into an order.
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.announce(ctx, domain.RoutingKeyProductDeleted, domain.ProductDeletedEvent{ProductID: id})

	return nil
}

func productEvent(p *domain.Product) domain.ProductEvent {
	return domain.ProductEvent{
		ProductID: p.ID,
		Name:      p.Name,
		Price:     p.Price,
		Stock:     p.Stock,
		Status:    p.Status,
	}
}

// announce broadcasts a domain event payload onto the shared exchange.
func (s *ProductService) announce(ctx context.Context, routingKey string, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("WARNING: could not encode %s event: %v", routingKey, err)
		return
	}

	if err := s.publisher.Publish(ctx, domain.OrderExchange, routingKey, body); err != nil {
		log.Printf("WARNING: could not publish %s event: %v", routingKey, err)
	}
}

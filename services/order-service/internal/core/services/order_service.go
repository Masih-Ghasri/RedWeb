package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/ports"
	"github.com/google/uuid"
)

type OrderService struct {
	repo  ports.OrderRepository
	cache ports.InventoryCache
}

func NewOrderService(repo ports.OrderRepository, cache ports.InventoryCache) *OrderService {
	return &OrderService{
		repo:  repo,
		cache: cache,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, cmd ports.CreateOrderCommand) (*domain.Order, error) {
	var orderItems []domain.OrderItem

	for _, itemCmd := range cmd.Items {
		stock, price, err := s.cache.GetProductStockAndPrice(ctx, itemCmd.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product %s check failed: %w", itemCmd.ProductID, err)
		}

		if stock < itemCmd.Quantity {
			return nil, domain.ErrInsufficientStock
		}

		orderItems = append(orderItems, domain.OrderItem{
			ProductID: itemCmd.ProductID,
			Quantity:  itemCmd.Quantity,
			Price:     price,
		})
	}

	orderID := uuid.New().String()
	order, err := domain.NewOrder(orderID, cmd.CustomerID, orderItems)
	if err != nil {
		return nil, err
	}

	eventPayload := domain.OrderCreatedEvent{
		OrderID:     order.ID,
		CustomerID:  order.CustomerID,
		TotalAmount: order.TotalAmount,
		Status:      string(order.Status),
	}

	payloadBytes, _ := json.Marshal(eventPayload)

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

	return order, nil
}

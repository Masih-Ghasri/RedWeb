package domain_test

import (
	"errors"
	"testing"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/domain"
)

func TestNewOrder(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		customerID  string
		items       []domain.OrderItem
		wantErr     error
		wantTotal   float64
		wantStatus  domain.OrderStatus
	}{
		{
			name:       "valid single item",
			id:         "ord-1",
			customerID: "alice",
			items: []domain.OrderItem{
				{ProductID: "prod-1", Quantity: 2, Price: 15.0},
			},
			wantErr:    nil,
			wantTotal:  30.0,
			wantStatus: domain.StatusPending,
		},
		{
			name:       "valid multiple items",
			id:         "ord-2",
			customerID: "bob",
			items: []domain.OrderItem{
				{ProductID: "prod-1", Quantity: 1, Price: 20.0},
				{ProductID: "prod-2", Quantity: 3, Price: 10.0},
			},
			wantErr:    nil,
			wantTotal:  50.0,
			wantStatus: domain.StatusPending,
		},
		{
			name:       "empty items slice",
			id:         "ord-3",
			customerID: "charlie",
			items:      []domain.OrderItem{},
			wantErr:    domain.ErrEmptyOrderItems,
		},
		{
			name:       "nil items",
			id:         "ord-4",
			customerID: "charlie",
			items:      nil,
			wantErr:    domain.ErrEmptyOrderItems,
		},
		{
			name:       "zero total amount",
			id:         "ord-5",
			customerID: "david",
			items: []domain.OrderItem{
				{ProductID: "prod-free", Quantity: 5, Price: 0.0},
			},
			wantErr: domain.ErrInvalidOrderAmount,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			order, err := domain.NewOrder(tc.id, tc.customerID, tc.items)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if order.ID != tc.id {
				t.Errorf("expected ID %s, got %s", tc.id, order.ID)
			}
			if order.CustomerID != tc.customerID {
				t.Errorf("expected customerID %s, got %s", tc.customerID, order.CustomerID)
			}
			if order.TotalAmount != tc.wantTotal {
				t.Errorf("expected total %f, got %f", tc.wantTotal, order.TotalAmount)
			}
			if order.Status != tc.wantStatus {
				t.Errorf("expected status %s, got %s", tc.wantStatus, order.Status)
			}
			if order.CreatedAt.IsZero() || order.UpdatedAt.IsZero() {
				t.Errorf("expected timestamps to be set")
			}
		})
	}
}

func TestOrder_CalculateTotal(t *testing.T) {
	order := &domain.Order{
		Items: []domain.OrderItem{
			{ProductID: "p1", Quantity: 2, Price: 12.5},
			{ProductID: "p2", Quantity: 4, Price: 5.0},
		},
	}

	order.CalculateTotal()
	expected := (2 * 12.5) + (4 * 5.0) // 25 + 20 = 45.0
	if order.TotalAmount != expected {
		t.Errorf("expected total %f, got %f", expected, order.TotalAmount)
	}
}

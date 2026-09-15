package domain_test

import (
	"errors"
	"testing"

	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/domain"
)

func TestNewProduct(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		pName     string
		desc      string
		price     float64
		stock     int
		wantErr   error
		wantState string
	}{
		{
			name:      "valid product",
			id:        "prod-1",
			pName:     "Coffee Mug",
			desc:      "Ceramic mug 350ml",
			price:     14.99,
			stock:     10,
			wantErr:   nil,
			wantState: domain.StatusActive,
		},
		{
			name:    "invalid price zero",
			id:      "prod-2",
			pName:   "Free Item",
			desc:    "Should not be zero",
			price:   0,
			stock:   5,
			wantErr: domain.ErrInvalidPrice,
		},
		{
			name:    "invalid price negative",
			id:      "prod-3",
			pName:   "Negative Item",
			desc:    "Negative price",
			price:   -10.0,
			stock:   5,
			wantErr: domain.ErrInvalidPrice,
		},
		{
			name:    "invalid stock zero on creation",
			id:      "prod-4",
			pName:   "Empty Item",
			desc:    "New products must have stock > 0",
			price:   20.0,
			stock:   0,
			wantErr: domain.ErrStockRequired,
		},
		{
			name:    "invalid stock negative",
			id:      "prod-5",
			pName:   "Negative Stock Item",
			desc:    "Negative stock is invalid",
			price:   20.0,
			stock:   -2,
			wantErr: domain.ErrStockRequired,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := domain.NewProduct(tc.id, tc.pName, tc.desc, tc.price, tc.stock)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.ID != tc.id {
				t.Errorf("expected ID %s, got %s", tc.id, p.ID)
			}
			if p.Status != tc.wantState {
				t.Errorf("expected Status %s, got %s", tc.wantState, p.Status)
			}
			if p.CreatedAt.IsZero() || p.UpdatedAt.IsZero() {
				t.Errorf("expected timestamps to be populated")
			}
		})
	}
}

func TestProduct_Update(t *testing.T) {
	p, err := domain.NewProduct("p1", "Item", "Desc", 10.0, 5)
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	t.Run("valid update with stock > 0", func(t *testing.T) {
		err := p.Update("Updated Item", "New Desc", 12.5, 8)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name != "Updated Item" || p.Price != 12.5 || p.Stock != 8 {
			t.Errorf("fields were not updated properly")
		}
		if p.Status != domain.StatusActive {
			t.Errorf("expected status %s, got %s", domain.StatusActive, p.Status)
		}
	})

	t.Run("update stock to 0 sets out of stock", func(t *testing.T) {
		err := p.Update("Updated Item", "New Desc", 12.5, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Stock != 0 {
			t.Errorf("expected stock 0, got %d", p.Stock)
		}
		if p.Status != domain.StatusOutOfStock {
			t.Errorf("expected status %s, got %s", domain.StatusOutOfStock, p.Status)
		}
	})

	t.Run("negative price is rejected", func(t *testing.T) {
		err := p.Update("Item", "Desc", -5.0, 10)
		if !errors.Is(err, domain.ErrInvalidPrice) {
			t.Fatalf("expected ErrInvalidPrice, got %v", err)
		}
	})

	t.Run("negative stock is rejected", func(t *testing.T) {
		err := p.Update("Item", "Desc", 10.0, -1)
		if !errors.Is(err, domain.ErrInvalidStock) {
			t.Fatalf("expected ErrInvalidStock, got %v", err)
		}
	})
}

func TestProduct_DecreaseStock(t *testing.T) {
	p, err := domain.NewProduct("p1", "Item", "Desc", 10.0, 5)
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	t.Run("decrease stock partially", func(t *testing.T) {
		err := p.DecreaseStock(3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Stock != 2 {
			t.Errorf("expected stock 2, got %d", p.Stock)
		}
		if p.Status != domain.StatusActive {
			t.Errorf("expected status %s, got %s", domain.StatusActive, p.Status)
		}
	})

	t.Run("decrease stock to zero", func(t *testing.T) {
		err := p.DecreaseStock(2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Stock != 0 {
			t.Errorf("expected stock 0, got %d", p.Stock)
		}
		if p.Status != domain.StatusOutOfStock {
			t.Errorf("expected status %s, got %s", domain.StatusOutOfStock, p.Status)
		}
	})

	t.Run("insufficient stock error", func(t *testing.T) {
		err := p.DecreaseStock(1)
		if !errors.Is(err, domain.ErrInsufficientStock) {
			t.Fatalf("expected ErrInsufficientStock, got %v", err)
		}
	})

	t.Run("invalid quantity zero or negative", func(t *testing.T) {
		if err := p.DecreaseStock(0); !errors.Is(err, domain.ErrInvalidQuantity) {
			t.Errorf("expected ErrInvalidQuantity for 0, got %v", err)
		}
		if err := p.DecreaseStock(-2); !errors.Is(err, domain.ErrInvalidQuantity) {
			t.Errorf("expected ErrInvalidQuantity for -2, got %v", err)
		}
	})
}

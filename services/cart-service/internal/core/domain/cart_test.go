package domain_test

import (
	"errors"
	"testing"

	"github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/core/domain"
)

func TestNewCart(t *testing.T) {
	cart := domain.NewCart("cust-100")
	if cart.CustomerID != "cust-100" {
		t.Errorf("expected customerID cust-100, got %s", cart.CustomerID)
	}
	if !cart.IsEmpty() {
		t.Errorf("new cart should be empty")
	}
	if cart.Total != 0 {
		t.Errorf("new cart total should be 0, got %f", cart.Total)
	}
	if cart.UpdatedAt.IsZero() {
		t.Errorf("updatedAt should not be zero")
	}
}

func TestCart_AddItem(t *testing.T) {
	cart := domain.NewCart("cust-100")

	t.Run("add first item", func(t *testing.T) {
		err := cart.AddItem("prod-1", 2, 25.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cart.IsEmpty() {
			t.Errorf("cart should not be empty")
		}
		if len(cart.Items) != 1 {
			t.Errorf("expected 1 item, got %d", len(cart.Items))
		}
		if cart.QuantityOf("prod-1") != 2 {
			t.Errorf("expected quantity 2, got %d", cart.QuantityOf("prod-1"))
		}
		if cart.Total != 50.0 {
			t.Errorf("expected total 50.0, got %f", cart.Total)
		}
	})

	t.Run("add same item merges quantity and refreshes price", func(t *testing.T) {
		err := cart.AddItem("prod-1", 3, 20.0) // price changed from 25 to 20
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cart.Items) != 1 {
			t.Errorf("items should be merged, got length %d", len(cart.Items))
		}
		if cart.QuantityOf("prod-1") != 5 {
			t.Errorf("expected merged quantity 5, got %d", cart.QuantityOf("prod-1"))
		}
		if cart.Total != 100.0 { // 5 * 20.0
			t.Errorf("expected total 100.0, got %f", cart.Total)
		}
	})

	t.Run("add second distinct item", func(t *testing.T) {
		err := cart.AddItem("prod-2", 1, 15.5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cart.Items) != 2 {
			t.Errorf("expected 2 items, got %d", len(cart.Items))
		}
		if cart.QuantityOf("prod-2") != 1 {
			t.Errorf("expected quantity 1, got %d", cart.QuantityOf("prod-2"))
		}
		if cart.Total != 115.5 { // 100.0 + 15.5
			t.Errorf("expected total 115.5, got %f", cart.Total)
		}
	})

	t.Run("invalid quantity rejected", func(t *testing.T) {
		if err := cart.AddItem("prod-3", 0, 10.0); !errors.Is(err, domain.ErrInvalidQuantity) {
			t.Errorf("expected ErrInvalidQuantity for 0, got %v", err)
		}
		if err := cart.AddItem("prod-3", -1, 10.0); !errors.Is(err, domain.ErrInvalidQuantity) {
			t.Errorf("expected ErrInvalidQuantity for -1, got %v", err)
		}
	})
}

func TestCart_UpdateQuantity(t *testing.T) {
	cart := domain.NewCart("cust-100")
	_ = cart.AddItem("prod-1", 3, 10.0)
	_ = cart.AddItem("prod-2", 2, 20.0)

	t.Run("update existing item quantity", func(t *testing.T) {
		err := cart.UpdateQuantity("prod-1", 5, 12.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cart.QuantityOf("prod-1") != 5 {
			t.Errorf("expected quantity 5, got %d", cart.QuantityOf("prod-1"))
		}
		// total = (5 * 12.0) + (2 * 20.0) = 60 + 40 = 100.0
		if cart.Total != 100.0 {
			t.Errorf("expected total 100.0, got %f", cart.Total)
		}
	})

	t.Run("update quantity to 0 removes the item", func(t *testing.T) {
		err := cart.UpdateQuantity("prod-1", 0, 12.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cart.QuantityOf("prod-1") != 0 {
			t.Errorf("expected item to be removed")
		}
		if len(cart.Items) != 1 {
			t.Errorf("expected 1 item left, got %d", len(cart.Items))
		}
		if cart.Total != 40.0 {
			t.Errorf("expected total 40.0, got %f", cart.Total)
		}
	})

	t.Run("negative quantity returns ErrInvalidQuantity", func(t *testing.T) {
		err := cart.UpdateQuantity("prod-2", -3, 20.0)
		if !errors.Is(err, domain.ErrInvalidQuantity) {
			t.Errorf("expected ErrInvalidQuantity, got %v", err)
		}
	})

	t.Run("updating absent product returns ErrProductNotInCart", func(t *testing.T) {
		err := cart.UpdateQuantity("nonexistent", 2, 10.0)
		if !errors.Is(err, domain.ErrProductNotInCart) {
			t.Errorf("expected ErrProductNotInCart, got %v", err)
		}
	})
}

func TestCart_RemoveItem(t *testing.T) {
	cart := domain.NewCart("cust-100")
	_ = cart.AddItem("prod-1", 2, 10.0)
	_ = cart.AddItem("prod-2", 1, 20.0)

	t.Run("remove existing item", func(t *testing.T) {
		err := cart.RemoveItem("prod-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cart.QuantityOf("prod-1") != 0 {
			t.Errorf("expected item to be removed")
		}
		if len(cart.Items) != 1 {
			t.Errorf("expected 1 item, got %d", len(cart.Items))
		}
		if cart.Total != 20.0 {
			t.Errorf("expected total 20.0, got %f", cart.Total)
		}
	})

	t.Run("remove missing item returns error", func(t *testing.T) {
		err := cart.RemoveItem("prod-1")
		if !errors.Is(err, domain.ErrProductNotInCart) {
			t.Errorf("expected ErrProductNotInCart, got %v", err)
		}
	})
}

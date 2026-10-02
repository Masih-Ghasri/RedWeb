package ports

import (
	"context"

	"github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/core/domain"
)

type AddItemCommand struct {
	CustomerID string
	ProductID  string
	Quantity   int
}

type UpdateItemCommand struct {
	CustomerID string
	ProductID  string
	Quantity   int
}

type CartUseCase interface {
	// GetCart returns an empty cart for a customer who has never added
	// anything — it never reports "not found".
	GetCart(ctx context.Context, customerID string) (*domain.Cart, error)

	// AddItem adds quantity to a line, merging a repeated product into the
	// existing line rather than creating a second one.
	AddItem(ctx context.Context, cmd AddItemCommand) (*domain.Cart, error)

	// UpdateItemQuantity sets an absolute quantity; zero removes the line.
	UpdateItemQuantity(ctx context.Context, cmd UpdateItemCommand) (*domain.Cart, error)

	// RemoveItem deletes one line from the cart.
	RemoveItem(ctx context.Context, customerID, productID string) (*domain.Cart, error)

	// ClearCart empties the cart. Removing a cart that does not exist
	// succeeds.
	ClearCart(ctx context.Context, customerID string) error
}

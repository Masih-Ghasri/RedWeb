package services

import (
	"context"
	"log"

	"github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/core/ports"
)

type CartService struct {
	repo   ports.CartRepository
	reader ports.ProductInfoReader
}

func NewCartService(repo ports.CartRepository, reader ports.ProductInfoReader) *CartService {
	return &CartService{
		repo:   repo,
		reader: reader,
	}
}

// GetCart returns the customer's cart, creating an empty one in memory when
// none is stored.
//
// Prices are refreshed from the shared product cache on the way out so a
// cart never displays a stale price. A product that has since been deleted
// keeps its last stored price rather than failing the whole read — the line
// is removed when the customer touches it again.
//
// The refreshed prices are not written back: a read should not turn into a
// write, and checkout re-reads them from the cache anyway.
func (s *CartService) GetCart(ctx context.Context, customerID string) (*domain.Cart, error) {
	cart, err := s.load(ctx, customerID)
	if err != nil {
		return nil, err
	}

	refreshed := false
	for i := range cart.Items {
		_, price, err := s.reader.GetProductInfo(ctx, cart.Items[i].ProductID)
		if err != nil {
			log.Printf("cart %s: could not refresh price of %s: %v", customerID, cart.Items[i].ProductID, err)
			continue
		}
		if price != cart.Items[i].Price {
			cart.Items[i].Price = price
			refreshed = true
		}
	}
	if refreshed {
		cart.Recalculate()
	}

	return cart, nil
}

// AddItem adds quantity to a cart line.
//
// The product is resolved before the cart is loaded, so an unknown product
// fails without doing any work, and the resulting quantity is checked
// against live stock — this is the first place in the system where adding
// more than exists is refused.
func (s *CartService) AddItem(ctx context.Context, cmd ports.AddItemCommand) (*domain.Cart, error) {
	if cmd.Quantity <= 0 {
		return nil, domain.ErrInvalidQuantity
	}

	stock, price, err := s.reader.GetProductInfo(ctx, cmd.ProductID)
	if err != nil {
		return nil, err
	}

	cart, err := s.load(ctx, cmd.CustomerID)
	if err != nil {
		return nil, err
	}

	if cart.QuantityOf(cmd.ProductID)+cmd.Quantity > stock {
		return nil, domain.ErrInsufficientStock
	}

	if err := cart.AddItem(cmd.ProductID, cmd.Quantity, price); err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, cart); err != nil {
		return nil, err
	}

	return cart, nil
}

// UpdateItemQuantity sets an absolute quantity for a line. Zero removes it.
//
// Quantity is absolute rather than a delta, so the stock check is a simple
// comparison against the live figure.
func (s *CartService) UpdateItemQuantity(ctx context.Context, cmd ports.UpdateItemCommand) (*domain.Cart, error) {
	if cmd.Quantity < 0 {
		return nil, domain.ErrInvalidQuantity
	}

	cart, err := s.load(ctx, cmd.CustomerID)
	if err != nil {
		return nil, err
	}

	// Zero means "remove", which needs no stock lookup — and if the line is
	// already gone the domain reports ErrProductNotInCart rather than
	// silently succeeding, so the caller learns the line never existed.
	if cmd.Quantity == 0 {
		if err := cart.UpdateQuantity(cmd.ProductID, 0, 0); err != nil {
			return nil, err
		}
		if err := s.repo.Save(ctx, cart); err != nil {
			return nil, err
		}
		return cart, nil
	}

	stock, price, err := s.reader.GetProductInfo(ctx, cmd.ProductID)
	if err != nil {
		return nil, err
	}

	if cmd.Quantity > stock {
		return nil, domain.ErrInsufficientStock
	}

	if err := cart.UpdateQuantity(cmd.ProductID, cmd.Quantity, price); err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, cart); err != nil {
		return nil, err
	}

	return cart, nil
}

// RemoveItem deletes one line from the cart.
func (s *CartService) RemoveItem(ctx context.Context, customerID, productID string) (*domain.Cart, error) {
	cart, err := s.load(ctx, customerID)
	if err != nil {
		return nil, err
	}

	if err := cart.RemoveItem(productID); err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, cart); err != nil {
		return nil, err
	}

	return cart, nil
}

// ClearCart removes the cart entirely. Deleting one that was never created
// succeeds, so it can be retried safely.
func (s *CartService) ClearCart(ctx context.Context, customerID string) error {
	return s.repo.Delete(ctx, customerID)
}

// load fetches a cart or synthesises an empty one, so callers never have to
// handle the "not created yet" case themselves.
func (s *CartService) load(ctx context.Context, customerID string) (*domain.Cart, error) {
	cart, err := s.repo.Get(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if cart == nil {
		return domain.NewCart(customerID), nil
	}
	return cart, nil
}

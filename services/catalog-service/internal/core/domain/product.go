package domain

import "time"

// Product lifecycle states persisted alongside every product.
const (
	StatusActive     = "ACTIVE"
	StatusOutOfStock = "OUT_OF_STOCK"
)

// Pagination bounds. The service layer clamps to these before querying and
// the persistence adapter falls back on them, so the two can never drift and
// a single request can never turn into an unbounded scan.
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type Product struct {
	ID          string
	Name        string
	Description string
	Price       float64
	Stock       int
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ProductFilter narrows and pages a product listing.
//
// It lives in domain rather than ports so the persistence adapter can
// implement FindAll without importing the ports package — the convention the
// other adapters in this repo already follow. Zero values mean "no
// restriction"; Page/PageSize are defaulted by the service layer, not here.
type ProductFilter struct {
	Q           string // case-insensitive substring of the name
	MinPrice    *float64
	MaxPrice    *float64
	InStockOnly bool
	Page        int // 1-based
	PageSize    int
}

// NewProduct validates and builds a product that is in stock by definition.
//
// All three audit fields (Status, CreatedAt, UpdatedAt) are set here rather
// than at the persistence layer, so an event published right after the
// product is created carries real values instead of zero values.
func NewProduct(id, name, desc string, price float64, stock int) (*Product, error) {
	if price <= 0 {
		return nil, ErrInvalidPrice
	}
	if stock <= 0 {
		return nil, ErrStockRequired
	}

	now := time.Now().UTC()
	return &Product{
		ID:          id,
		Name:        name,
		Description: desc,
		Price:       price,
		Stock:       stock,
		Status:      StatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update applies new values to an existing product and refreshes UpdatedAt.
//
// Unlike NewProduct, a stock of zero is legal here — marking a product out
// of stock is a normal operation — but a negative stock never is.
func (p *Product) Update(name, desc string, price float64, stock int) error {
	if price <= 0 {
		return ErrInvalidPrice
	}
	if stock < 0 {
		return ErrInvalidStock
	}

	p.Name = name
	p.Description = desc
	p.Price = price
	p.Stock = stock
	p.Status = statusForStock(stock)
	p.UpdatedAt = time.Now().UTC()

	return nil
}

// DecreaseStock removes quantity from the current stock, refusing to go
// below zero. It is reserved for order-time stock reservation, which is not
// implemented yet — CRUD stock changes go through Update instead.
func (p *Product) DecreaseStock(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	if p.Stock < quantity {
		return ErrInsufficientStock
	}

	p.Stock -= quantity
	p.Status = statusForStock(p.Stock)
	p.UpdatedAt = time.Now().UTC()

	return nil
}

func statusForStock(stock int) string {
	if stock == 0 {
		return StatusOutOfStock
	}
	return StatusActive
}

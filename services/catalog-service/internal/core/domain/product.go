package domain

import (
	"errors"
	"time"
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

func (p *Product) DecreaseStock(quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}
	if p.Stock < quantity {
		return errors.New("insufficient stock")
	}

	p.Stock -= quantity
	if p.Stock == 0 {
		p.Status = "OUT_OF_STOCK"
	}
	p.UpdatedAt = time.Now()

	return nil
}

func NewProduct(id, name, desc string, price float64, stock int) (*Product, error) {
	if price <= 0 {
		return nil, errors.New("price must be greater than zero")
	}
	if stock <= 0 {
		return nil, errors.New("stock must be greater than zero")
	}
	return &Product{
		ID:          id,
		Name:        name,
		Description: desc,
		Price:       price,
		Stock:       stock}, nil
}

package domain

import "time"

// CartItem is one line of a cart.
//
// Price is a snapshot taken when the item was last written, refreshed from
// the shared product cache on every read. It exists so a cart renders with a
// number instead of having to re-price the world on display.
type CartItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// Cart is a customer's shopping cart.
//
// The struct doubles as the Redis wire format (key "cart:{customerID}", a
// single JSON blob), which is why every field carries a json tag. Renaming a
// tag silently orphans previously stored carts.
type Cart struct {
	CustomerID string     `json:"customer_id"`
	Items      []CartItem `json:"items"`
	Total      float64    `json:"total"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// NewCart returns an empty cart. An empty cart is a normal state, not an
// error — GET on a customer who has never added anything answers 200.
func NewCart(customerID string) *Cart {
	return &Cart{
		CustomerID: customerID,
		Items:      []CartItem{},
		UpdatedAt:  time.Now().UTC(),
	}
}

// IsEmpty reports whether the cart has no line items.
func (c *Cart) IsEmpty() bool {
	return len(c.Items) == 0
}

// QuantityOf returns how many units of a product are already in the cart,
// or 0 when the product is absent. Callers use it to check whether adding
// more would exceed the available stock.
func (c *Cart) QuantityOf(productID string) int {
	if i := c.IndexOf(productID); i >= 0 {
		return c.Items[i].Quantity
	}
	return 0
}

// IndexOf returns the position of a product in the cart, or -1 if absent.
func (c *Cart) IndexOf(productID string) int {
	for i := range c.Items {
		if c.Items[i].ProductID == productID {
			return i
		}
	}
	return -1
}

// AddItem appends a product, or increases the quantity of one already in the
// cart. Adding the same product twice merges the quantities instead of
// creating a second line.
func (c *Cart) AddItem(productID string, quantity int, price float64) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	if i := c.IndexOf(productID); i >= 0 {
		c.Items[i].Quantity += quantity
		c.Items[i].Price = price // adopt the current price on every write
	} else {
		c.Items = append(c.Items, CartItem{
			ProductID: productID,
			Quantity:  quantity,
			Price:     price,
		})
	}

	c.Recalculate()
	return nil
}

// UpdateQuantity replaces the quantity of a product already in the cart.
// Passing zero removes the line entirely, which is what makes
// `PUT …/items/{id}` with `{"quantity":0}` equivalent to a DELETE.
func (c *Cart) UpdateQuantity(productID string, quantity int, price float64) error {
	if quantity < 0 {
		return ErrInvalidQuantity
	}

	i := c.IndexOf(productID)
	if i < 0 {
		return ErrProductNotInCart
	}

	if quantity == 0 {
		c.removeAt(i)
		return nil
	}

	c.Items[i].Quantity = quantity
	c.Items[i].Price = price

	c.Recalculate()
	return nil
}

// RemoveItem deletes a line from the cart.
func (c *Cart) RemoveItem(productID string) error {
	i := c.IndexOf(productID)
	if i < 0 {
		return ErrProductNotInCart
	}

	c.removeAt(i)
	return nil
}

// Recalculate recomputes the cart total amount and updates the timestamp.
func (c *Cart) Recalculate() {
	var total float64
	for _, item := range c.Items {
		total += item.Price * float64(item.Quantity)
	}
	c.Total = total
	c.UpdatedAt = time.Now().UTC()
}

func (c *Cart) removeAt(i int) {
	c.Items = append(c.Items[:i], c.Items[i+1:]...)
	c.Recalculate()
}

package dto

// AddItemRequest is the body of POST /api/v1/carts/{customerID}/items.
//
// Quantity uses gt=0 rather than "required,gt=0": for an int, `required`
// already rejects zero, and stating the real constraint directly is clearer.
// An absent field decodes to 0 and is rejected the same way.
type AddItemRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"gt=0"`
}

// UpdateItemRequest is the body of PUT /api/v1/carts/{customerID}/items/{productID}.
//
// Quantity is absolute (not a delta) and zero is meaningful — it removes the
// line. It is a *pointer* so "quantity omitted" and "quantity: 0" stay
// distinguishable: without this, a client that forgets the field would
// silently delete the line instead of being told to fix its request.
type UpdateItemRequest struct {
	Quantity *int `json:"quantity" validate:"required,gte=0"`
}

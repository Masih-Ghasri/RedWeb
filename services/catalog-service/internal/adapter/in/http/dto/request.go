package dto

// CreateProductRequest defines the payload for creating a new product.
type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"gt=0"`
	Stock       int     `json:"stock" validate:"gt=0"`
}

// UpdateProductRequest defines the payload for updating an existing product.
type UpdateProductRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"gt=0"`
	Stock       int     `json:"stock" validate:"gte=0"`
}

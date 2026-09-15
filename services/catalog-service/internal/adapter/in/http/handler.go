package http

import (
	"encoding/json"
	"net/http"

	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/adapter/in/http/dto"
	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/ports"
)

type ProductHandler struct {
	useCase ports.ProductUseCase
}

func NewProductHandler(uc ports.ProductUseCase) *ProductHandler {
	return &ProductHandler{useCase: uc}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	// TODO: Call pkg/validator here

	cmd := ports.CreateProductCommand{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}

	product, err := h.useCase.CreateProduct(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.FromDomain(product))
}

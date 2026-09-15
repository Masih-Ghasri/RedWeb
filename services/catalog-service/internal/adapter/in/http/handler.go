package http

import (
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

// Health handles basic liveness checks.
func (h *ProductHandler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ListProducts handles GET /api/v1/products.
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	page, err := intParam(params.Get("page"), 0)
	if err != nil {
		writeBadQuery(w, "page", err)
		return
	}

	pageSize, err := intParam(params.Get("page_size"), 0)
	if err != nil {
		writeBadQuery(w, "page_size", err)
		return
	}

	minPrice, err := floatParam(params.Get("min_price"))
	if err != nil {
		writeBadQuery(w, "min_price", err)
		return
	}

	maxPrice, err := floatParam(params.Get("max_price"))
	if err != nil {
		writeBadQuery(w, "max_price", err)
		return
	}

	if minPrice != nil && maxPrice != nil && *minPrice > *maxPrice {
		http.Error(w, "min_price cannot be greater than max_price", http.StatusBadRequest)
		return
	}

	inStock, err := boolParam(params.Get("in_stock"))
	if err != nil {
		writeBadQuery(w, "in_stock", err)
		return
	}

	list, err := h.useCase.ListProducts(r.Context(), ports.ListProductsQuery{
		Q:           params.Get("q"),
		MinPrice:    minPrice,
		MaxPrice:    maxPrice,
		InStockOnly: inStock,
		Page:        page,
		PageSize:    pageSize,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.FromProductList(list))
}

// GetProduct handles GET /api/v1/products/{id}.
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "product id is required", http.StatusBadRequest)
		return
	}

	product, err := h.useCase.GetProduct(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.FromDomain(product))
}

// CreateProduct handles POST /api/v1/products.
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProductRequest
	if !decodeAndValidate(w, r, &req) {
		return
	}

	product, err := h.useCase.CreateProduct(r.Context(), ports.CreateProductCommand{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, dto.FromDomain(product))
}

// UpdateProduct handles PUT /api/v1/products/{id}.
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "product id is required", http.StatusBadRequest)
		return
	}

	var req dto.UpdateProductRequest
	if !decodeAndValidate(w, r, &req) {
		return
	}

	product, err := h.useCase.UpdateProduct(r.Context(), id, ports.UpdateProductCommand{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.FromDomain(product))
}

// DeleteProduct handles DELETE /api/v1/products/{id}. 204 carries no body, so
// there is nothing to encode on the success path.
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "product id is required", http.StatusBadRequest)
		return
	}

	if err := h.useCase.DeleteProduct(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

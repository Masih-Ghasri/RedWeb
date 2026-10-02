package http

import (
	"net/http"

	"github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/adapter/in/http/dto"
	"github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/core/ports"
)

type CartHandler struct {
	useCase ports.CartUseCase
}

func NewCartHandler(uc ports.CartUseCase) *CartHandler {
	return &CartHandler{useCase: uc}
}

// Health handles basic liveness checks.
func (h *CartHandler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GetCart handles GET /api/v1/carts/{customerID}.
// Returns the customer's shopping cart or an empty cart if not yet initialized.
func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	customerID, ok := pathID(w, r, "customerID")
	if !ok {
		return
	}

	cart, err := h.useCase.GetCart(r.Context(), customerID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.FromDomain(cart))
}

// AddItem handles POST /api/v1/carts/{customerID}/items.
func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	customerID, ok := pathID(w, r, "customerID")
	if !ok {
		return
	}

	var req dto.AddItemRequest
	if !decodeAndValidate(w, r, &req) {
		return
	}

	cart, err := h.useCase.AddItem(r.Context(), ports.AddItemCommand{
		CustomerID: customerID,
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.FromDomain(cart))
}

// UpdateItem handles PUT /api/v1/carts/{customerID}/items/{productID}.
func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	customerID, ok := pathID(w, r, "customerID")
	if !ok {
		return
	}
	productID, ok := pathID(w, r, "productID")
	if !ok {
		return
	}

	var req dto.UpdateItemRequest
	if !decodeAndValidate(w, r, &req) {
		return
	}

	cart, err := h.useCase.UpdateItemQuantity(r.Context(), ports.UpdateItemCommand{
		CustomerID: customerID,
		ProductID:  productID,
		Quantity:   *req.Quantity,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.FromDomain(cart))
}

// RemoveItem handles DELETE /api/v1/carts/{customerID}/items/{productID}.
//
// It answers 404 when the line is not there rather than succeeding, so a
// caller who targets the wrong product learns about it.
func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	customerID, ok := pathID(w, r, "customerID")
	if !ok {
		return
	}
	productID, ok := pathID(w, r, "productID")
	if !ok {
		return
	}

	cart, err := h.useCase.RemoveItem(r.Context(), customerID, productID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.FromDomain(cart))
}

// ClearCart handles DELETE /api/v1/carts/{customerID} and returns the now
// empty cart, so a client can render the cleared state without a second
// request. Clearing a cart that does not exist also succeeds.
func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	customerID, ok := pathID(w, r, "customerID")
	if !ok {
		return
	}

	if err := h.useCase.ClearCart(r.Context(), customerID); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.FromDomain(domain.NewCart(customerID)))
}

// pathID reads a path wildcard, answering 400 when it is empty.
//
// Go's mux can hand back an empty value for a collapsed path segment, and an
// empty id would otherwise turn into a Redis key like "cart:".
func pathID(w http.ResponseWriter, r *http.Request, name string) (string, bool) {
	value := r.PathValue(name)
	if value == "" {
		http.Error(w, name+" is required", http.StatusBadRequest)
		return "", false
	}
	return value, true
}

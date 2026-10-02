package http

import (
	"net/http"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/adapter/in/http/dto"
	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/ports"
)

type OrderHandler struct {
	useCase ports.OrderUseCase
}

func NewOrderHandler(uc ports.OrderUseCase) *OrderHandler {
	return &OrderHandler{useCase: uc}
}

// Health handles basic liveness checks.
func (h *OrderHandler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// CreateOrder handles POST /api/v1/orders.
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateOrderRequest
	if !decodeAndValidate(w, r, &req) {
		return
	}

	if req.CartID == "" && req.CustomerID == "" && len(req.Items) == 0 {
		http.Error(w, "cart_id or customer_id with items required", http.StatusBadRequest)
		return
	}

	cmd := ports.CreateOrderCommand{
		CartID:     req.CartID,
		CustomerID: req.CustomerID,
	}
	for _, item := range req.Items {
		cmd.Items = append(cmd.Items, ports.CreateOrderItemCommand{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	order, err := h.useCase.CreateOrder(r.Context(), cmd)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, dto.FromDomain(order))
}

// GetOrder handles GET /api/v1/orders/{id}.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := pathValue(w, r, "id")
	if id == "" {
		return
	}

	order, err := h.useCase.GetOrder(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.FromDomain(order))
}

// ListOrders handles GET /api/v1/orders?customer_id=&status=&page=&page_size=
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	page, err := intParam(params.Get("page"), 0)
	if err != nil {
		http.Error(w, `invalid value for "page"`, http.StatusBadRequest)
		return
	}

	pageSize, err := intParam(params.Get("page_size"), 0)
	if err != nil {
		http.Error(w, `invalid value for "page_size"`, http.StatusBadRequest)
		return
	}

	list, err := h.useCase.ListOrders(r.Context(), ports.ListOrdersQuery{
		CustomerID: params.Get("customer_id"),
		Status:     params.Get("status"),
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.FromOrderList(list))
}

// CancelOrder handles POST /api/v1/orders/{id}/cancel.
//
// POST rather than DELETE: cancelling is a state transition on a resource
// that still exists afterwards, and DELETE would imply the order is removed.
func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	id := pathValue(w, r, "id")
	if id == "" {
		return
	}

	order, err := h.useCase.CancelOrder(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.FromDomain(order))
}

// pathValue reads a path wildcard, answering 400 when it is empty. A "" is
// returned alongside the response so callers can bail out uniformly.
func pathValue(w http.ResponseWriter, r *http.Request, name string) string {
	value := r.PathValue(name)
	if value == "" {
		http.Error(w, name+" is required", http.StatusBadRequest)
		return ""
	}
	return value
}

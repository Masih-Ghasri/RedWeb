package http

import (
	"encoding/json"
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

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd := ports.CreateOrderCommand{
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
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	res := dto.OrderResponse{
		ID:          order.ID,
		TotalAmount: order.TotalAmount,
		Status:      string(order.Status),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

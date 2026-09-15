package http

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *OrderHandler) {
	mux.HandleFunc("POST /api/v1/orders", handler.CreateOrder)
}

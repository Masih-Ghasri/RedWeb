package http

import "net/http"

// RegisterRoutes wires every order route onto the mux.
//
// Go 1.22+ method-and-wildcard patterns mean an unmatched method on a known
// path returns 405 rather than falling through to the catch-all 404. No two
// patterns below can match the same request — they differ by method or by
// segment count — so the mux will not panic at registration time.
func RegisterRoutes(mux *http.ServeMux, handler *OrderHandler) {
	mux.HandleFunc("GET /healthz", handler.Health)

	mux.HandleFunc("GET /api/v1/orders", handler.ListOrders)
	mux.HandleFunc("POST /api/v1/orders", handler.CreateOrder)
	mux.HandleFunc("GET /api/v1/orders/{id}", handler.GetOrder)
	mux.HandleFunc("POST /api/v1/orders/{id}/cancel", handler.CancelOrder)
}

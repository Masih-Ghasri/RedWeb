package http

import "net/http"

// RegisterRoutes wires every cart route onto the mux.
//
// Go 1.22+ method-and-wildcard patterns mean an unmatched method on a known
// path returns 405 rather than falling through to the catch-all 404, and no
// two patterns here can match the same request (they differ by method or by
// segment count), so the mux will not panic at registration time.
func RegisterRoutes(mux *http.ServeMux, handler *CartHandler) {
	mux.HandleFunc("GET /healthz", handler.Health)

	mux.HandleFunc("GET /api/v1/carts/{customerID}", handler.GetCart)
	mux.HandleFunc("POST /api/v1/carts/{customerID}/items", handler.AddItem)
	mux.HandleFunc("PUT /api/v1/carts/{customerID}/items/{productID}", handler.UpdateItem)
	mux.HandleFunc("DELETE /api/v1/carts/{customerID}/items/{productID}", handler.RemoveItem)
	mux.HandleFunc("DELETE /api/v1/carts/{customerID}", handler.ClearCart)
}

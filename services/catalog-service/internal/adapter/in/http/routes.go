package http

import "net/http"

// RegisterRoutes wires every catalog route onto the mux.
//
// Go 1.22+ method-and-wildcard patterns mean an unmatched method on a known
// path returns 405 rather than falling through to the catch-all 404.
func RegisterRoutes(mux *http.ServeMux, handler *ProductHandler) {
	mux.HandleFunc("GET /healthz", handler.Health)

	mux.HandleFunc("GET /api/v1/products", handler.ListProducts)
	mux.HandleFunc("GET /api/v1/products/{id}", handler.GetProduct)
	mux.HandleFunc("POST /api/v1/products", handler.CreateProduct)
	mux.HandleFunc("PUT /api/v1/products/{id}", handler.UpdateProduct)
	mux.HandleFunc("DELETE /api/v1/products/{id}", handler.DeleteProduct)
}

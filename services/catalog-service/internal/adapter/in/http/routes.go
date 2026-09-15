package http

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *ProductHandler) {
	mux.HandleFunc("POST /api/v1/products", handler.CreateProduct)
	// mux.HandleFunc("GET /api/v1/products/{id}", handler.GetProduct)
}

package http

import (
	"errors"
	"log"
	"net/http"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/ports"
)

// writeError maps domain and port errors to appropriate HTTP status codes.
func writeError(w http.ResponseWriter, err error) {
	switch {
	// 404 — the addressed thing does not exist
	case errors.Is(err, domain.ErrOrderNotFound),
		errors.Is(err, domain.ErrProductNotFound),
		errors.Is(err, ports.ErrCartNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)

	// 400 — the request itself is malformed or empty
	case errors.Is(err, domain.ErrEmptyOrderItems),
		errors.Is(err, domain.ErrInvalidOrderAmount):
		http.Error(w, err.Error(), http.StatusBadRequest)

	// 409 — well-formed, but conflicts with current state
	case errors.Is(err, domain.ErrInsufficientStock),
		errors.Is(err, domain.ErrOrderNotCancelable):
		http.Error(w, err.Error(), http.StatusConflict)

	// 503 — a dependency is down; this is *not* the client's fault and must
	// never be reported as 4xx or as a generic validation failure
	case errors.Is(err, ports.ErrCartUnavailable):
		http.Error(w, "cart service is unavailable, please retry", http.StatusServiceUnavailable)

	default:
		log.Printf("order: unhandled error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

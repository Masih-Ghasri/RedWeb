package http

import (
	"errors"
	"log"
	"net/http"

	"github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/core/domain"
)

// writeError maps a use-case error onto an HTTP response.
//
// Domain failures the caller can act on receive their proper 4xx code.
// Everything else is an infrastructure fault (Redis down, malformed stored
// cart) and is answered generically — the raw error is logged instead of
// being handed to the client, because it can contain connection details.
func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrProductNotFound),
		errors.Is(err, domain.ErrProductNotInCart):
		http.Error(w, err.Error(), http.StatusNotFound)

	case errors.Is(err, domain.ErrInvalidQuantity):
		http.Error(w, err.Error(), http.StatusBadRequest)

	case errors.Is(err, domain.ErrInsufficientStock):
		http.Error(w, err.Error(), http.StatusConflict)

	default:
		log.Printf("cart: unhandled error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

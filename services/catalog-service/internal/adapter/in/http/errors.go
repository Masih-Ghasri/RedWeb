package http

import (
	"errors"
	"log"
	"net/http"

	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/domain"
)

// writeError maps a use-case error onto an HTTP response.
//
// Domain failures the caller can act on receive their proper 4xx code. Every
// other error is an infrastructure fault (Mongo down, network timeout) and is
// answered with a generic message — the raw error string is logged instead of
// being leaked to the client, because it can contain connection details.
func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)

	case errors.Is(err, domain.ErrInvalidPrice),
		errors.Is(err, domain.ErrStockRequired),
		errors.Is(err, domain.ErrInvalidStock),
		errors.Is(err, domain.ErrInvalidQuantity):
		http.Error(w, err.Error(), http.StatusBadRequest)

	case errors.Is(err, domain.ErrInsufficientStock):
		http.Error(w, err.Error(), http.StatusConflict)

	default:
		log.Printf("catalog: unhandled error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

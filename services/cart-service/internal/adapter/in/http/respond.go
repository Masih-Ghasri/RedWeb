package http

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// maxBodyBytes caps how much of a request body is read. Without it a client
// could hold a connection open by streaming an arbitrarily large payload;
// once the cap is hit the decoder fails and the handler answers 400.
const maxBodyBytes = 1 << 20 // 1 MiB

// validate is the shared struct validator instance.
// Source: https://pkg.go.dev/github.com/go-playground/validator/v10#WithRequiredStructEnabled
var validate = validator.New(validator.WithRequiredStructEnabled())

// decodeBody decodes a JSON request body into dst, capped at maxBodyBytes.
func decodeBody(r *http.Request, dst any) error {
	return json.NewDecoder(io.LimitReader(r.Body, maxBodyBytes)).Decode(dst)
}

// decodeAndValidate decodes and validates a JSON request body.
func decodeAndValidate(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := decodeBody(r, dst); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return false
	}
	if err := validate.Struct(dst); err != nil {
		http.Error(w, "validation failed: "+err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("cart: failed to encode response: %v", err)
	}
}

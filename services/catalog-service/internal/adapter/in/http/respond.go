package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// maxBodyBytes caps how much of a request body is read. Without it a client
// could hold a connection open by streaming an arbitrarily large payload.
const maxBodyBytes = 1 << 20 // 1 MiB

// validate is the shared struct validator instance.
// Source: https://pkg.go.dev/github.com/go-playground/validator/v10#WithRequiredStructEnabled
var validate = validator.New(validator.WithRequiredStructEnabled())

// decodeBody reads a JSON request body into dst, refusing anything larger
// than maxBodyBytes.
func decodeBody(r *http.Request, dst any) error {
	return json.NewDecoder(io.LimitReader(r.Body, maxBodyBytes)).Decode(dst)
}

// decodeAndValidate reads a JSON request body capped at maxBodyBytes (1 MiB)
// and validates struct constraints using github.com/go-playground/validator/v10.
// Source: https://pkg.go.dev/github.com/go-playground/validator/v10
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

// writeJSON writes a JSON response with the given status.
//
// Encoding failures are logged rather than turned into a second response —
// once WriteHeader has run the status line is already on the wire, so a
// follow-up error would only corrupt the body.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("catalog: failed to encode response: %v", err)
	}
}

// writeBadQuery reports an unparseable query parameter as a client error.
func writeBadQuery(w http.ResponseWriter, param string, err error) {
	http.Error(w, fmt.Sprintf("invalid value for %q: %v", param, err), http.StatusBadRequest)
}

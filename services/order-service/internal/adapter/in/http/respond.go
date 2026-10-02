package http

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

// maxBodyBytes caps how much of a request body is read. Without it a client
// could hold a connection open by streaming an arbitrarily large payload;
// once the cap is hit the decoder fails and the handler answers 400.
const maxBodyBytes = 1 << 20 // 1 MiB

// validate is the shared struct validator instance.
// Source: https://pkg.go.dev/github.com/go-playground/validator/v10#WithRequiredStructEnabled
var validate = validator.New(validator.WithRequiredStructEnabled())

// writeJSON writes a JSON response with the given status.
//
// Encoding failures are logged rather than turned into a second response —
// once WriteHeader has run the status line is already on the wire, so a
// follow-up error would only corrupt the body.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("order: failed to encode response: %v", err)
	}
}

// decodeBody reads a JSON request body into dst, refusing anything larger
// than maxBodyBytes. Kept identical to the cart and catalog helpers so the
// three handlers read the same way.
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

// intParam parses an optional integer query parameter. An absent value yields
// fallback (so the service can apply its own default) while a present-but
// unparsable value is an error — silently ignoring it would make the caller
// believe their filter was applied when it was not.
func intParam(raw string, fallback int) (int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, err
	}
	return value, nil
}

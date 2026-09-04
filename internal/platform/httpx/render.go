package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
)

// MaxBodyBytes bounds request bodies. A larger body is a 400, never a 500.
const MaxBodyBytes = 1 << 20 // 1 MiB

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// WriteJSON writes v as a JSON response with the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) { writeJSON(w, status, v) }

// DecodeJSON reads exactly one JSON object from the request body into dst, with a
// size limit and unknown-field rejection. A malformed, oversized, or trailing-data
// body returns a 400 APIError.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return NewError(http.StatusBadRequest, CodeBadRequest, "Request body too large")
		}
		return NewError(http.StatusBadRequest, CodeBadRequest, "Malformed JSON body")
	}

	if dec.More() {
		return NewError(http.StatusBadRequest, CodeBadRequest, "Request body must contain a single JSON object")
	}
	return nil
}

package httpx

import (
	"errors"
	"log/slog"
	"net/http"
)

// ErrorCode is the stable, documented machine-readable error identifier (ADR-0002).
type ErrorCode string

// The documented error code vocabulary. These are labels returned to clients, not
// secrets.
const (
	CodeValidation         ErrorCode = "VALIDATION_ERROR"
	CodeBadRequest         ErrorCode = "BAD_REQUEST"
	CodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS" //nolint:gosec // G101: error code label, not a credential
	CodeEmailTaken         ErrorCode = "EMAIL_ALREADY_REGISTERED"
	CodeUnauthenticated    ErrorCode = "UNAUTHENTICATED"
	CodeForbidden          ErrorCode = "FORBIDDEN"
	CodeNotFound           ErrorCode = "NOT_FOUND"
	CodeConflict           ErrorCode = "CONFLICT"
	CodeRateLimited        ErrorCode = "RATE_LIMITED"
	CodeInternal           ErrorCode = "INTERNAL"
)

// APIError is an error that carries an HTTP status and the client-facing envelope
// fields. Handlers return it (directly or wrapped) and the router writes it.
type APIError struct {
	Status  int
	Code    ErrorCode
	Message string
	Fields  map[string]string // only for validation failures
}

func (e *APIError) Error() string { return string(e.Code) + ": " + e.Message }

// NewError builds a simple APIError with no field detail.
func NewError(status int, code ErrorCode, message string) *APIError {
	return &APIError{Status: status, Code: code, Message: message}
}

// ValidationError builds a 400 carrying per-field messages.
func ValidationError(fields map[string]string) *APIError {
	return &APIError{
		Status:  http.StatusBadRequest,
		Code:    CodeValidation,
		Message: "Request validation failed",
		Fields:  fields,
	}
}

type errorBody struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code    ErrorCode         `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// WriteError writes err as the standard error envelope. Any status >= 500 is logged
// server-side with the request id and returned to the client as a generic INTERNAL
// error with no implementation detail (ADR-0002).
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		apiErr = NewError(http.StatusInternalServerError, CodeInternal, "An internal error occurred")
	}

	if apiErr.Status >= 500 {
		slog.ErrorContext(r.Context(), "request failed",
			slog.String("request_id", RequestID(r.Context())),
			slog.String("error", err.Error()),
		)
		apiErr = NewError(apiErr.Status, CodeInternal, "An internal error occurred")
	}

	writeJSON(w, apiErr.Status, errorBody{Error: errorPayload{
		Code:    apiErr.Code,
		Message: apiErr.Message,
		Fields:  apiErr.Fields,
	}})
}

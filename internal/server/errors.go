package server

import (
	"encoding/json"
	"net/http"
)

// ErrorEnvelope is the body shape returned on every 4xx/5xx response.
type ErrorEnvelope struct {
	ErrorType string `json:"error_type"`
	Retriable bool   `json:"retriable"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// Each typed error has a stable HTTP status + retriable flag.
type errSpec struct {
	status    int
	retriable bool
}

var errCatalog = map[string]errSpec{
	"JWT_MISSING":              {http.StatusUnauthorized, false},
	"JWT_VALIDATION_FAILED":    {http.StatusUnauthorized, false},
	"PERMISSION_DENIED":        {http.StatusForbidden, false},
	"UNKNOWN_TOOL":             {http.StatusNotFound, false},
	"NOT_FOUND":                {http.StatusNotFound, false},
	"SCHEMA_VALIDATION_FAILED": {http.StatusBadRequest, false},
	"BAD_REQUEST":              {http.StatusBadRequest, false},
	"BACKEND_UNAVAILABLE":      {http.StatusServiceUnavailable, true},
	"QUERY_TIMEOUT":            {http.StatusGatewayTimeout, true},
	"BACKEND_ERROR":            {http.StatusInternalServerError, true},
	"INTERNAL_ERROR":           {http.StatusInternalServerError, true},
	"FORBIDDEN_CALLER":         {http.StatusForbidden, false},
}

// WriteError writes the envelope with the appropriate status code.
func WriteError(w http.ResponseWriter, r *http.Request, errType, msg string) {
	spec, ok := errCatalog[errType]
	if !ok {
		spec = errSpec{http.StatusInternalServerError, true}
		errType = "INTERNAL_ERROR"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(spec.status)
	_ = json.NewEncoder(w).Encode(ErrorEnvelope{
		ErrorType: errType,
		Retriable: spec.retriable,
		Message:   msg,
		RequestID: r.Header.Get("X-Request-ID"),
	})
}

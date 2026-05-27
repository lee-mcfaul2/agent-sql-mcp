package server

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Trace extracts the inbound W3C `traceparent` and starts a span per request.
// Without this each MCP call would be a separate root trace; pairing the
// gateway's propagator with this middleware stitches the MCP work into the
// caller's trace -- a Tempo search by trace_id finds gateway + sandbox +
// MCP spans for the same prompt.
func Trace(next http.Handler) http.Handler {
	tracer := otel.Tracer("agent-sql-mcp")
	prop := otel.GetTextMapPropagator()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := prop.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := tracer.Start(ctx, r.URL.Path)
		defer span.End()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestID assigns or propagates X-Request-ID.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-ID", id)
		r.Header.Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

// Recover catches panics in handlers and emits INTERNAL_ERROR.
func Recover(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic in handler",
						"err", rec, "stack", string(debug.Stack()), "path", r.URL.Path)
					WriteError(w, r, "INTERNAL_ERROR", "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// AccessLog emits one structured log line per request.
func AccessLog(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", w.Header().Get("X-Request-ID"),
			)
		})
	}
}

// SPIFFECheck rejects callers whose forwarded identity doesn't match the
// configured gateway identity. We accept either Linkerd's `l5d-client-id`
// (DNS-form mesh identity, e.g. `<sa>.<ns>.serviceaccount.identity.linkerd.cluster.local`)
// or Envoy/Istio's `X-Forwarded-Client-Cert` (SPIFFE URI in the cert).
// The configured `expectedIdentity` may be a substring of either form;
// the demo uses the Linkerd DNS form because Linkerd does not emit XFCC.
func SPIFFECheck(expectedIdentity string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip the SPIFFE check for the local liveness probes.
			if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" || r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}
			xfcc := r.Header.Get("X-Forwarded-Client-Cert")
			l5d := r.Header.Get("l5d-client-id")
			ok := (xfcc != "" && strings.Contains(xfcc, expectedIdentity)) ||
				(l5d != "" && strings.Contains(l5d, expectedIdentity))
			if !ok {
				WriteError(w, r, "FORBIDDEN_CALLER", "caller identity mismatch")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

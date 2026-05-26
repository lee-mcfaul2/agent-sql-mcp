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
)

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

// SPIFFECheck rejects callers whose forwarded SPIFFE doesn't match the configured gateway identity.
// Linkerd injects the caller cert into X-Forwarded-Client-Cert; we parse the SPIFFE URI from there.
func SPIFFECheck(expectedSPIFFE string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip the SPIFFE check for the local liveness probes.
			if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" || r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}
			hdr := r.Header.Get("X-Forwarded-Client-Cert")
			ok := strings.Contains(hdr, expectedSPIFFE)
			// REVERT-BEFORE-RELEASE: unsafe verbose debug for SPIFFE mismatch hunt
			slog.Default().Info("spiffe.check.unsafe_debug",
				"path", r.URL.Path,
				"expected_spiffe", expectedSPIFFE,
				"xfcc_header", hdr,
				"l5d_client_id", r.Header.Get("l5d-client-id"),
				"match", ok,
			)
			if !ok {
				WriteError(w, r, "FORBIDDEN_CALLER", "caller SPIFFE mismatch")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/auth"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/obs"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/schemas"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
)

type Deps struct {
	Log            *slog.Logger
	Pool           store.Pool
	JWT            *auth.Validator
	Validators     *schemas.Validators
	Catalog        *schemas.Catalog
	SchemaVersion  string
	ExpectedSPIFFE string
	QueryTimeout   func() (sec int)
}

func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(RequestID)
	r.Use(Trace)
	r.Use(Recover(deps.Log))
	r.Use(AccessLog(deps.Log))
	r.Use(SPIFFECheck(deps.ExpectedSPIFFE))

	r.Get("/healthz", handleHealthz)
	r.Get("/readyz", handleReadyz(deps))
	r.Method("GET", "/metrics", obs.MetricsHandler())

	r.Get("/handshake", handleHandshake(deps))
	r.Post("/v1/tools/{tool}", handleTool(deps))

	return r
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func handleReadyz(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := deps.Pool.Acquire(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		conn.Release()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}

func handleHandshake(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		obs.HandshakeCallsTotal.Inc()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"schema_version": deps.SchemaVersion,
			"schema_digest":  "sha256:" + deps.Catalog.Digest,
		})
	}
}


package obs

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var Registry = prometheus.NewRegistry()

var (
	ToolCallsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "sqlmcp_tool_calls_total", Help: "Inbound tool calls."},
		[]string{"tool", "outcome"},
	)
	ToolDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "sqlmcp_tool_duration_seconds", Help: "Tool latency.", Buckets: prometheus.DefBuckets},
		[]string{"tool", "outcome"},
	)
	AuthzDenialsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "sqlmcp_authz_denials_total", Help: "Per-tool authz denials."},
		[]string{"tool", "reason"},
	)
	JWTFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "sqlmcp_jwt_failures_total", Help: "JWT validation failures by reason."},
		[]string{"reason"},
	)
	JWKSRefreshFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "sqlmcp_jwks_refresh_failures_total", Help: "JWKS refresh failures."},
		[]string{"reason"},
	)
	SchemaFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "sqlmcp_schema_failures_total", Help: "Schema validation failures."},
		[]string{"tool", "direction"},
	)
	PGPoolAcquireSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{Name: "sqlmcp_pg_pool_acquire_seconds", Help: "pgx pool acquire time.", Buckets: prometheus.DefBuckets},
	)
	PGErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "sqlmcp_pg_errors_total", Help: "Postgres errors by sqlstate bucket."},
		[]string{"sqlstate"},
	)
	HandshakeCallsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{Name: "sqlmcp_handshake_calls_total", Help: "Handshake calls received."},
	)
	SchemasLoaded = prometheus.NewGauge(
		prometheus.GaugeOpts{Name: "sqlmcp_schemas_loaded", Help: "Count of tool schemas loaded at startup."},
	)
	AuditWritesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "sqlmcp_audit_writes_total", Help: "Audit log writes by outcome."},
		[]string{"outcome"},
	)
)

func init() {
	Registry.MustRegister(
		ToolCallsTotal, ToolDuration, AuthzDenialsTotal,
		JWTFailuresTotal, JWKSRefreshFailuresTotal,
		SchemaFailuresTotal, PGPoolAcquireSeconds, PGErrorsTotal,
		HandshakeCallsTotal, SchemasLoaded, AuditWritesTotal,
	)
}

func MetricsHandler() http.Handler {
	return promhttp.HandlerFor(Registry, promhttp.HandlerOpts{Registry: Registry})
}

package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// HTTP
	ListenAddr string
	// OIDC
	OIDCIssuer          string
	OIDCAudience        string
	JWKSRefreshSeconds  int
	// Belt-and-braces SPIFFE check
	GatewaySPIFFE string
	// Postgres
	DatabaseURL         string
	DBPoolMax           int
	QueryTimeoutSeconds int
	// Schema version (build-time const elsewhere; env override for dev)
	SchemaVersion string
	// Telemetry
	OTLPEndpoint string
	LogLevel     string
	ServiceName  string
}

func Load() (*Config, error) {
	c := &Config{
		ListenAddr:          getenv("SQLMCP_LISTEN_ADDR", ":8443"),
		OIDCIssuer:          os.Getenv("SQLMCP_OIDC_ISSUER"),
		OIDCAudience:        os.Getenv("SQLMCP_OIDC_AUDIENCE"),
		JWKSRefreshSeconds:  getint("SQLMCP_JWKS_REFRESH_SECONDS", 3600),
		GatewaySPIFFE:       os.Getenv("SQLMCP_GATEWAY_SPIFFE"),
		DatabaseURL:         os.Getenv("SQLMCP_DATABASE_URL"),
		DBPoolMax:           getint("SQLMCP_DB_POOL_MAX", 25),
		QueryTimeoutSeconds: getint("SQLMCP_QUERY_TIMEOUT_SECONDS", 5),
		SchemaVersion:       getenv("SQLMCP_SCHEMA_VERSION", "v1"),
		OTLPEndpoint:        os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		LogLevel:            getenv("SQLMCP_LOG_LEVEL", "info"),
		ServiceName:         getenv("SQLMCP_SERVICE_NAME", "agent-sql-mcp"),
	}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) validate() error {
	required := map[string]string{
		"SQLMCP_OIDC_ISSUER":    c.OIDCIssuer,
		"SQLMCP_OIDC_AUDIENCE":  c.OIDCAudience,
		"SQLMCP_GATEWAY_SPIFFE": c.GatewaySPIFFE,
		"SQLMCP_DATABASE_URL":   c.DatabaseURL,
	}
	for k, v := range required {
		if v == "" {
			return fmt.Errorf("required env var not set: %s", k)
		}
	}
	if c.DBPoolMax < 1 {
		return fmt.Errorf("SQLMCP_DB_POOL_MAX must be >= 1, got %d", c.DBPoolMax)
	}
	if c.QueryTimeoutSeconds < 1 {
		return fmt.Errorf("SQLMCP_QUERY_TIMEOUT_SECONDS must be >= 1, got %d", c.QueryTimeoutSeconds)
	}
	return nil
}

func (c *Config) QueryTimeout() time.Duration {
	return time.Duration(c.QueryTimeoutSeconds) * time.Second
}

func (c *Config) JWKSRefresh() time.Duration {
	return time.Duration(c.JWKSRefreshSeconds) * time.Second
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getint(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

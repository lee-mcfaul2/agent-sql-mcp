package config

import (
	"testing"
)

func required(t *testing.T) {
	t.Helper()
	t.Setenv("SQLMCP_OIDC_ISSUER", "https://idp.example")
	t.Setenv("SQLMCP_OIDC_AUDIENCE", "agent-sql-mcp")
	t.Setenv("SQLMCP_GATEWAY_SPIFFE", "spiffe://x/gateway")
	t.Setenv("SQLMCP_DATABASE_URL", "postgres://x/db")
}

func TestLoad_Defaults(t *testing.T) {
	required(t)
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.ListenAddr != ":8443" {
		t.Errorf("ListenAddr default: %q", c.ListenAddr)
	}
	if c.DBPoolMax != 25 {
		t.Errorf("DBPoolMax default: %d", c.DBPoolMax)
	}
}

func TestLoad_MissingRequired(t *testing.T) {
	t.Setenv("SQLMCP_OIDC_ISSUER", "")
	t.Setenv("SQLMCP_OIDC_AUDIENCE", "agent-sql-mcp")
	t.Setenv("SQLMCP_GATEWAY_SPIFFE", "spiffe://x/gateway")
	t.Setenv("SQLMCP_DATABASE_URL", "postgres://x/db")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when SQLMCP_OIDC_ISSUER is empty")
	}
}

func TestLoad_BadInt(t *testing.T) {
	required(t)
	t.Setenv("SQLMCP_DB_POOL_MAX", "0")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when DBPoolMax=0")
	}
}

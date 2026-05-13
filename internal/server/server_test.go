package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
)

type stubPool struct{}

func (s *stubPool) Acquire(ctx context.Context) (store.Conn, error) { return &stubConn{}, nil }
func (s *stubPool) Close()                                          {}

type stubConn struct{}

func (c *stubConn) Exec(_ context.Context, _ string, _ ...any) (any, error) { return nil, nil }
func (c *stubConn) Query(_ context.Context, _ string, _ ...any) (store.Rows, error) {
	return nil, nil
}
func (c *stubConn) QueryRow(_ context.Context, _ string, _ ...any) store.Row { return nil }
func (c *stubConn) Release()                                                 {}

func nullLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestHealthz(t *testing.T) {
	r := NewRouter(Deps{
		Log:            nullLogger(),
		Pool:           &stubPool{},
		ExpectedSPIFFE: "spiffe://x/gateway",
	})
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status: %d", w.Code)
	}
}

func TestSPIFFE_Rejects(t *testing.T) {
	r := NewRouter(Deps{
		Log:            nullLogger(),
		Pool:           &stubPool{},
		ExpectedSPIFFE: "spiffe://x/gateway",
	})
	req := httptest.NewRequest("POST", "/v1/tools/search_customer", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
	var env ErrorEnvelope
	_ = json.NewDecoder(w.Body).Decode(&env)
	if env.ErrorType != "FORBIDDEN_CALLER" {
		t.Errorf("error_type: %s", env.ErrorType)
	}
}

func TestErrorEnvelope_Catalog(t *testing.T) {
	cases := map[string]int{
		"JWT_MISSING":              401,
		"PERMISSION_DENIED":        403,
		"NOT_FOUND":                404,
		"SCHEMA_VALIDATION_FAILED": 400,
		"BACKEND_UNAVAILABLE":      503,
		"QUERY_TIMEOUT":            504,
	}
	for et, want := range cases {
		req := httptest.NewRequest("POST", "/x", nil)
		w := httptest.NewRecorder()
		WriteError(w, req, et, "msg")
		if w.Code != want {
			t.Errorf("%s -> %d, want %d", et, w.Code, want)
		}
	}
}

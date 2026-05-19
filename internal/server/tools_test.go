package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lee-mcfaul2/agent-sql-mcp/internal/auth"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/schemas"
)

func validSchemaBytes() map[string][]byte {
	m := map[string][]byte{}
	for _, t := range []string{"search_customer", "lookup_customer", "list_orders", "list_transactions", "get_order"} {
		m[t+".request.json"] = []byte(`{"type":"object"}`)
		m[t+".response.json"] = []byte(`{"type":"object"}`)
	}
	return m
}

func TestRoute_MissingJWT_Returns401(t *testing.T) {
	cat, _ := schemas.LoadFromBytes(validSchemaBytes())
	v, _ := schemas.CompileValidators(cat)
	r := NewRouter(Deps{
		Log:            nullLogger(),
		Pool:           &stubPool{},
		Validators:     v,
		Catalog:        cat,
		SchemaVersion:  "v1",
		ExpectedSPIFFE: "spiffe://x/gateway",
		QueryTimeout:   func() int { return 5 },
		JWT:            &auth.Validator{},
	})
	req := httptest.NewRequest("POST", "/v1/tools/search_customer", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("X-Forwarded-Client-Cert", "URI=spiffe://x/gateway")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status: %d", w.Code)
	}
	var env ErrorEnvelope
	_ = json.NewDecoder(w.Body).Decode(&env)
	if env.ErrorType != "JWT_MISSING" {
		t.Errorf("error_type: %s", env.ErrorType)
	}
}

// Silence unused import in case CI is picky.
var _ = time.Now
var _ context.Context = context.Background()

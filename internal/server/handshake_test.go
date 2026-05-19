package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lee-mcfaul2/agent-sql-mcp/internal/schemas"
)

func TestHandshake_ReturnsVersionAndDigest(t *testing.T) {
	cat, _ := schemas.LoadFromBytes(map[string][]byte{
		"search_customer.request.json":    []byte(`{"type":"object"}`),
		"search_customer.response.json":   []byte(`{"type":"object"}`),
		"lookup_customer.request.json":    []byte(`{"type":"object"}`),
		"lookup_customer.response.json":   []byte(`{"type":"object"}`),
		"list_orders.request.json":        []byte(`{"type":"object"}`),
		"list_orders.response.json":       []byte(`{"type":"object"}`),
		"list_transactions.request.json":  []byte(`{"type":"object"}`),
		"list_transactions.response.json": []byte(`{"type":"object"}`),
		"get_order.request.json":          []byte(`{"type":"object"}`),
		"get_order.response.json":         []byte(`{"type":"object"}`),
	})

	r := NewRouter(Deps{
		Log:            nullLogger(),
		Pool:           &stubPool{},
		Catalog:        cat,
		SchemaVersion:  "v1",
		ExpectedSPIFFE: "spiffe://x/gateway",
	})

	req := httptest.NewRequest("GET", "/handshake", nil)
	req.Header.Set("X-Forwarded-Client-Cert", "URI=spiffe://x/gateway")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status: %d", w.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["schema_version"] != "v1" {
		t.Errorf("version: %s", body["schema_version"])
	}
	if !strings.HasPrefix(body["schema_digest"], "sha256:") {
		t.Errorf("digest format: %s", body["schema_digest"])
	}
}

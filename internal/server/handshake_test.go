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
	files := map[string][]byte{}
	for _, t := range []string{
		"search_customer", "lookup_customer", "list_orders", "list_transactions", "get_order",
		"list_all_customers", "list_all_orders", "list_all_transactions",
	} {
		files[t+".request.json"] = []byte(`{"type":"object"}`)
		files[t+".response.json"] = []byte(`{"type":"object"}`)
	}
	cat, _ := schemas.LoadFromBytes(files)

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

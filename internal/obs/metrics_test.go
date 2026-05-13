package obs

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsHandler_ExposesCounter(t *testing.T) {
	ToolCallsTotal.WithLabelValues("search_customer", "ok").Inc()
	r := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	MetricsHandler().ServeHTTP(w, r)
	body := w.Body.String()
	if !strings.Contains(body, "sqlmcp_tool_calls_total") {
		t.Fatalf("metrics output missing counter: %s", body)
	}
}

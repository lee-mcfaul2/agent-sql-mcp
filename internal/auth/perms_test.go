package auth

import (
	"strings"
	"testing"
)

func TestToolPerms_AllToolsRegistered(t *testing.T) {
	expected := []string{
		"search_customer", "lookup_customer", "list_orders", "list_transactions", "get_order",
		"list_all_customers", "list_all_orders", "list_all_transactions",
	}
	for _, tool := range expected {
		if _, ok := ToolPerms[tool]; !ok {
			t.Errorf("missing perms for tool: %s", tool)
		}
	}
}

func TestToolPerms_RequiredForUnknown(t *testing.T) {
	_, err := RequiredFor("nope")
	if err == nil || !strings.Contains(err.Error(), "unknown tool") {
		t.Fatalf("expected unknown-tool error, got: %v", err)
	}
}

func TestHasAll_Matrix(t *testing.T) {
	cases := []struct {
		name     string
		have     []string
		required []string
		want     bool
	}{
		{"empty required passes", []string{}, []string{}, true},
		{"single match", []string{"customers:read"}, []string{"customers:read"}, true},
		{"missing one fails", []string{"orders:read"}, []string{"customers:read"}, false},
		{"superset passes", []string{"customers:read", "orders:read", "audit:read"}, []string{"customers:read"}, true},
		{"two required, one present", []string{"audit:read"}, []string{"audit:read", "audit:export"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := UserClaims{Permissions: tc.have}
			if got := c.HasAll(tc.required); got != tc.want {
				t.Errorf("HasAll: got %v want %v", got, tc.want)
			}
		})
	}
}

func TestMissing_ReportsExactlyWhatIsMissing(t *testing.T) {
	c := UserClaims{Permissions: []string{"orders:read"}}
	got := c.Missing([]string{"customers:read", "orders:read", "audit:read"})
	if len(got) != 2 {
		t.Fatalf("expected 2 missing, got %d: %v", len(got), got)
	}
}

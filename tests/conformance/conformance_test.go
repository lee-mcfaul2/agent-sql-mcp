//go:build integration

package conformance_test

import (
	"context"
	"strings"
	"testing"

	"github.com/lee-mcfaul2/agent-sql-mcp/internal/auth"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/tools"
)

var conformanceClaims = auth.UserClaims{Sub: "conformance", Permissions: []string{"customers:read"}}

func TestNoInjection_TrailingQuoteIsInert(t *testing.T) {
	p, cleanup := bootForConformance(t)
	defer cleanup()
	hostile := "'; DROP TABLE customers; --"
	_, err := tools.SearchCustomer(context.Background(), p, conformanceClaims, tools.SearchCustomerArgs{Name: &hostile})
	if err != nil {
		t.Errorf("hostile input produced an error (it shouldn't): %v", err)
	}
}

func TestNoInjection_BackslashesEscaped(t *testing.T) {
	p, cleanup := bootForConformance(t)
	defer cleanup()
	hostile := `\\'`
	_, err := tools.SearchCustomer(context.Background(), p, conformanceClaims, tools.SearchCustomerArgs{Name: &hostile})
	if err != nil {
		t.Errorf("backslash input errored: %v", err)
	}
}

func TestAudit_RowWrittenAfterCall(t *testing.T) {
	p, cleanup := bootForConformance(t)
	defer cleanup()
	err := store.WriteAudit(context.Background(), p, store.AuditEntry{
		UserSub: "alice", Tool: "search_customer", Outcome: "ok", DurationMs: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	c, _ := p.Acquire(context.Background())
	defer c.Release()
	row := c.QueryRow(context.Background(), "SELECT user_sub, tool, outcome FROM mcp_audit ORDER BY id DESC LIMIT 1")
	var sub, tool, outcome string
	if err := row.Scan(&sub, &tool, &outcome); err != nil {
		t.Fatal(err)
	}
	if sub != "alice" || tool != "search_customer" || outcome != "ok" {
		t.Errorf("audit row mismatch: %s/%s/%s", sub, tool, outcome)
	}
}

func TestSQLConstants_NoFormatVerbs(t *testing.T) {
	for _, sql := range []string{
		store.SQLSearchCustomer, store.SQLLookupCustomer,
		store.SQLListOrders, store.SQLListTransactions, store.SQLGetOrder, store.SQLGetOrderItems, store.SQLInsertAudit,
	} {
		if strings.Contains(sql, "%s") || strings.Contains(sql, "%v") {
			t.Errorf("format verb in SQL: %s", sql)
		}
	}
}

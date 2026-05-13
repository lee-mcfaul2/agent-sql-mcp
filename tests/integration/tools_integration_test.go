//go:build integration

package integration_test

import (
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/lee-mcfaul2/agent-sql-mcp/internal/server"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/tools"
)

func nullLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newPool(t *testing.T, dsn string) store.Pool {
	t.Helper()
	p, err := store.New(context.Background(), dsn, 8, 5)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Close)
	return p
}

func TestSearchCustomer_Live(t *testing.T) {
	dsn, cleanup := bootPostgres(t)
	defer cleanup()
	p := newPool(t, dsn)

	name := "Customer 0"
	res, err := tools.SearchCustomer(context.Background(), p, tools.SearchCustomerArgs{Name: &name})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Customers) == 0 {
		t.Errorf("no customers returned for name=%q", name)
	}
}

func TestLookupCustomer_LiveAndNotFound(t *testing.T) {
	dsn, cleanup := bootPostgres(t)
	defer cleanup()
	p := newPool(t, dsn)

	res, err := tools.LookupCustomer(context.Background(), p, tools.LookupCustomerArgs{CustomerID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if res.Customer.ID != 1 {
		t.Errorf("id: %d", res.Customer.ID)
	}

	_, err = tools.LookupCustomer(context.Background(), p, tools.LookupCustomerArgs{CustomerID: 99999})
	if err != tools.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListOrders_Live(t *testing.T) {
	dsn, cleanup := bootPostgres(t)
	defer cleanup()
	p := newPool(t, dsn)

	res, err := tools.ListOrders(context.Background(), p, tools.ListOrdersArgs{CustomerID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Orders) == 0 {
		t.Errorf("no orders for customer 1 (seed should ensure at least one)")
	}
}

func TestGetOrder_LiveWithItems(t *testing.T) {
	dsn, cleanup := bootPostgres(t)
	defer cleanup()
	p := newPool(t, dsn)

	res, err := tools.GetOrder(context.Background(), p, tools.GetOrderArgs{OrderID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if res.Order.ID != 1 {
		t.Errorf("id: %d", res.Order.ID)
	}
	if len(res.Order.LineItems) == 0 {
		t.Errorf("expected at least one line item for order 1")
	}
}

func TestReadyz_Live(t *testing.T) {
	dsn, cleanup := bootPostgres(t)
	defer cleanup()
	p := newPool(t, dsn)

	r := server.NewRouter(server.Deps{
		Log:            nullLogger(),
		Pool:           p,
		ExpectedSPIFFE: "spiffe://x/gateway",
		QueryTimeout:   func() int { return 5 },
	})
	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("readyz status: %d", w.Code)
	}
}

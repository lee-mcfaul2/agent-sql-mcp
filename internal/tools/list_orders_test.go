package tools

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
)

func TestListOrders_HappyPath(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cols := []string{"id", "customer_id", "status", "total_cents", "currency", "placed_at"}
	id := int64(42)
	// Params now: (customer_id *bigint, since *timestamptz, canSeeAtlantis bool, limit int)
	mock.ExpectQuery(`SELECT .* FROM orders`).
		WithArgs(&id, (*time.Time)(nil), false, 25).
		WillReturnRows(pgxmock.NewRows(cols).
			AddRow(int64(1), int64(42), "paid", int64(1000), "USD", now))

	res, err := ListOrders(context.Background(), &adaptPool{mock: mock}, basicReadClaims, ListOrdersArgs{CustomerID: &id})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Orders) != 1 || res.Orders[0].Status != "paid" {
		t.Errorf("unexpected rows: %+v", res.Orders)
	}
}

func TestListOrders_LimitRespected(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	limit := 5
	id := int64(42)
	cols := []string{"id", "customer_id", "status", "total_cents", "currency", "placed_at"}
	mock.ExpectQuery(`SELECT`).
		WithArgs(&id, (*time.Time)(nil), false, limit).
		WillReturnRows(pgxmock.NewRows(cols))
	_, err := ListOrders(context.Background(), &adaptPool{mock: mock}, basicReadClaims, ListOrdersArgs{CustomerID: &id, Limit: &limit})
	if err != nil {
		t.Fatal(err)
	}
}

// TestListOrders_NoCustomerID asserts the cross-customer browse path: when
// customer_id is omitted the handler still calls the query (with nil) so the
// SQL's "$1::bigint IS NULL OR ..." branch lists across all customers,
// subject to the row-level Atlantis filter on the JOINed customers row.
func TestListOrders_NoCustomerID(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	cols := []string{"id", "customer_id", "status", "total_cents", "currency", "placed_at"}
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	// CustomerID nil → first pgx arg is (*int64)(nil). canSeeAtlantis=false.
	mock.ExpectQuery(`JOIN customers`).
		WithArgs((*int64)(nil), (*time.Time)(nil), false, 25).
		WillReturnRows(pgxmock.NewRows(cols).
			AddRow(int64(1), int64(7), "paid", int64(1000), "USD", now))
	res, err := ListOrders(context.Background(), &adaptPool{mock: mock}, basicReadClaims, ListOrdersArgs{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Orders) != 1 {
		t.Errorf("expected 1 row, got %d", len(res.Orders))
	}
}

// TestListOrders_NoCustomerID_AtlantisVisibleToPrivileged asserts callers with
// customers:atlantis:read pass canSeeAtlantis=true so the SQL filter short-
// circuits and orders for Atlantis customers come through during a browse.
func TestListOrders_NoCustomerID_AtlantisVisibleToPrivileged(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	cols := []string{"id", "customer_id", "status", "total_cents", "currency", "placed_at"}
	mock.ExpectQuery(`JOIN customers`).
		WithArgs((*int64)(nil), (*time.Time)(nil), true, 25).
		WillReturnRows(pgxmock.NewRows(cols))
	_, err := ListOrders(context.Background(), &adaptPool{mock: mock}, atlantisReadClaims, ListOrdersArgs{})
	if err != nil {
		t.Fatal(err)
	}
}

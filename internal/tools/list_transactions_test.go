package tools

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
)

func TestListTransactions_HappyPath(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cols := []string{"id", "customer_id", "amount_cents", "kind", "created_at"}
	id := int64(42)
	// Params: (customer_id *bigint, since *timestamptz, canSeeAtlantis bool, limit int)
	mock.ExpectQuery(`SELECT .* FROM transactions`).
		WithArgs(&id, (*time.Time)(nil), false, 25).
		WillReturnRows(pgxmock.NewRows(cols).
			AddRow(int64(1), int64(42), int64(1000), "payment", now))

	res, err := ListTransactions(context.Background(), &adaptPool{mock: mock}, basicReadClaims, ListTransactionsArgs{CustomerID: &id})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Transactions) != 1 || res.Transactions[0].Kind != "payment" {
		t.Errorf("unexpected rows: %+v", res.Transactions)
	}
}

func TestListTransactions_LimitRespected(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	limit := 5
	id := int64(42)
	cols := []string{"id", "customer_id", "amount_cents", "kind", "created_at"}
	mock.ExpectQuery(`SELECT`).
		WithArgs(&id, (*time.Time)(nil), false, limit).
		WillReturnRows(pgxmock.NewRows(cols))
	_, err := ListTransactions(context.Background(), &adaptPool{mock: mock}, basicReadClaims, ListTransactionsArgs{CustomerID: &id, Limit: &limit})
	if err != nil {
		t.Fatal(err)
	}
}

// TestListTransactions_NoCustomerID asserts cross-customer browse: nil
// customer_id triggers the "$1::bigint IS NULL OR ..." branch and the
// JOINed Atlantis filter still applies.
func TestListTransactions_NoCustomerID(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	cols := []string{"id", "customer_id", "amount_cents", "kind", "created_at"}
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`JOIN customers`).
		WithArgs((*int64)(nil), (*time.Time)(nil), false, 25).
		WillReturnRows(pgxmock.NewRows(cols).
			AddRow(int64(1), int64(7), int64(1000), "payment", now))
	res, err := ListTransactions(context.Background(), &adaptPool{mock: mock}, basicReadClaims, ListTransactionsArgs{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Transactions) != 1 {
		t.Errorf("expected 1 row, got %d", len(res.Transactions))
	}
}

// TestListTransactions_NoCustomerID_AtlantisVisibleToPrivileged asserts the
// canSeeAtlantis=true path on a no-customer-id browse.
func TestListTransactions_NoCustomerID_AtlantisVisibleToPrivileged(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	cols := []string{"id", "customer_id", "amount_cents", "kind", "created_at"}
	mock.ExpectQuery(`JOIN customers`).
		WithArgs((*int64)(nil), (*time.Time)(nil), true, 25).
		WillReturnRows(pgxmock.NewRows(cols))
	_, err := ListTransactions(context.Background(), &adaptPool{mock: mock}, atlantisReadClaims, ListTransactionsArgs{})
	if err != nil {
		t.Fatal(err)
	}
}

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
	mock.ExpectQuery(`SELECT id, customer_id, amount_cents, kind, created_at`).
		WithArgs(int64(42), (*time.Time)(nil), 25).
		WillReturnRows(pgxmock.NewRows(cols).
			AddRow(int64(1), int64(42), int64(1000), "payment", now))

	res, err := ListTransactions(context.Background(), &adaptPool{mock: mock}, ListTransactionsArgs{CustomerID: 42})
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
	cols := []string{"id", "customer_id", "amount_cents", "kind", "created_at"}
	mock.ExpectQuery(`SELECT`).
		WithArgs(int64(42), (*time.Time)(nil), limit).
		WillReturnRows(pgxmock.NewRows(cols))
	_, err := ListTransactions(context.Background(), &adaptPool{mock: mock}, ListTransactionsArgs{CustomerID: 42, Limit: &limit})
	if err != nil {
		t.Fatal(err)
	}
}

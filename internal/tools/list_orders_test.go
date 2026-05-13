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
	mock.ExpectQuery(`SELECT id, customer_id, status, total_cents, currency, placed_at`).
		WithArgs(int64(42), (*time.Time)(nil), 25).
		WillReturnRows(pgxmock.NewRows(cols).
			AddRow(int64(1), int64(42), "paid", int64(1000), "USD", now))

	res, err := ListOrders(context.Background(), &adaptPool{mock: mock}, ListOrdersArgs{CustomerID: 42})
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
	cols := []string{"id", "customer_id", "status", "total_cents", "currency", "placed_at"}
	mock.ExpectQuery(`SELECT`).
		WithArgs(int64(42), (*time.Time)(nil), limit).
		WillReturnRows(pgxmock.NewRows(cols))
	_, err := ListOrders(context.Background(), &adaptPool{mock: mock}, ListOrdersArgs{CustomerID: 42, Limit: &limit})
	if err != nil {
		t.Fatal(err)
	}
}

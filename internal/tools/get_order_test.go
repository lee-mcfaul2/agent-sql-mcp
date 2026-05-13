package tools

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
)

func TestGetOrder_WithItems(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`FROM orders\s+WHERE id`).
		WithArgs(int64(7)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "customer_id", "status", "total_cents", "currency", "placed_at"}).
			AddRow(int64(7), int64(42), "paid", int64(2500), "USD", now))
	mock.ExpectQuery(`FROM order_items`).
		WithArgs(int64(7)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "sku", "quantity", "unit_cents"}).
			AddRow(int64(1), "ABC-1", 1, int64(1000)).
			AddRow(int64(2), "ABC-2", 3, int64(500)))

	res, err := GetOrder(context.Background(), &adaptPool{mock: mock}, GetOrderArgs{OrderID: 7})
	if err != nil {
		t.Fatal(err)
	}
	if res.Order.ID != 7 || len(res.Order.LineItems) != 2 {
		t.Errorf("unexpected: %+v", res.Order)
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	mock.ExpectQuery(`FROM orders\s+WHERE id`).
		WithArgs(int64(99)).
		WillReturnError(pgx.ErrNoRows)
	_, err := GetOrder(context.Background(), &adaptPool{mock: mock}, GetOrderArgs{OrderID: 99})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

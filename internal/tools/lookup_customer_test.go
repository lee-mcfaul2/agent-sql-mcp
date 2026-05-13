package tools

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
)

func TestLookupCustomer_Found(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cols := []string{"id", "name", "email", "phone", "address", "created_at"}
	mock.ExpectQuery(`SELECT id, name, email, phone, address, created_at`).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows(cols).
			AddRow(int64(42), "Bob", "b@x.com", (*string)(nil), (*string)(nil), now))

	res, err := LookupCustomer(context.Background(), &adaptPool{mock: mock}, LookupCustomerArgs{CustomerID: 42})
	if err != nil {
		t.Fatal(err)
	}
	if res.Customer.ID != 42 {
		t.Errorf("id: %d", res.Customer.ID)
	}
}

func TestLookupCustomer_NotFound(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	mock.ExpectQuery(`SELECT id, name, email, phone, address, created_at`).
		WithArgs(int64(999)).
		WillReturnError(pgx.ErrNoRows)

	_, err := LookupCustomer(context.Background(), &adaptPool{mock: mock}, LookupCustomerArgs{CustomerID: 999})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

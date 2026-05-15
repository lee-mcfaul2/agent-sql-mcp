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
	cols := []string{"id", "name", "email", "phone", "address", "created_at", "region"}
	mock.ExpectQuery(`SELECT id, name, email, phone, address, created_at, region`).
		WithArgs(int64(42), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows(cols).
			AddRow(int64(42), "Bob", "b@x.com", (*string)(nil), (*string)(nil), now, "north-america"))

	res, err := LookupCustomer(context.Background(), &adaptPool{mock: mock}, basicReadClaims, LookupCustomerArgs{CustomerID: 42})
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
	mock.ExpectQuery(`SELECT id, name, email, phone, address, created_at, region`).
		WithArgs(int64(999), pgxmock.AnyArg()).
		WillReturnError(pgx.ErrNoRows)

	_, err := LookupCustomer(context.Background(), &adaptPool{mock: mock}, basicReadClaims, LookupCustomerArgs{CustomerID: 999})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// TestLookupCustomer_AtlantisHiddenForNonPrivilegedUser asserts that a basic
// reader lookup-by-id of an atlantis customer returns ErrNotFound (because the
// row-level filter excludes it from the result set).
func TestLookupCustomer_AtlantisHiddenForNonPrivilegedUser(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	// canSeeAtlantis=false -> filter excludes atlantis row -> zero rows -> ErrNoRows.
	mock.ExpectQuery(`region != 'atlantis'`).
		WithArgs(int64(7), false).
		WillReturnError(pgx.ErrNoRows)

	_, err := LookupCustomer(context.Background(), &adaptPool{mock: mock}, basicReadClaims, LookupCustomerArgs{CustomerID: 7})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for hidden atlantis row, got %v", err)
	}
}

// TestLookupCustomer_AtlantisVisibleToPrivilegedUser asserts a caller with the
// atlantis perm can lookup the same atlantis customer by id.
func TestLookupCustomer_AtlantisVisibleToPrivilegedUser(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cols := []string{"id", "name", "email", "phone", "address", "created_at", "region"}
	mock.ExpectQuery(`SELECT .* FROM customers`).
		WithArgs(int64(7), true).
		WillReturnRows(pgxmock.NewRows(cols).
			AddRow(int64(7), "Atlantis Marine", "sec@atlantis.example", (*string)(nil), (*string)(nil), now, "atlantis"))

	res, err := LookupCustomer(context.Background(), &adaptPool{mock: mock}, atlantisReadClaims, LookupCustomerArgs{CustomerID: 7})
	if err != nil {
		t.Fatal(err)
	}
	if res.Customer.Region != "atlantis" {
		t.Errorf("expected region=atlantis, got %q", res.Customer.Region)
	}
}

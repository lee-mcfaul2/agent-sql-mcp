package tools

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
	"github.com/pashagolub/pgxmock/v4"
)

// adaptPool wraps a pgxmock connection as our store.Pool.
type adaptPool struct{ mock pgxmock.PgxPoolIface }

func (a *adaptPool) Acquire(ctx context.Context) (store.Conn, error) {
	return &adaptConn{mock: a.mock}, nil
}
func (a *adaptPool) Close() { a.mock.Close() }

type adaptConn struct{ mock pgxmock.PgxPoolIface }

func (c *adaptConn) Exec(ctx context.Context, sql string, args ...any) (any, error) {
	return c.mock.Exec(ctx, sql, args...)
}
func (c *adaptConn) Query(ctx context.Context, sql string, args ...any) (store.Rows, error) {
	r, err := c.mock.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &adaptRows{rows: r}, nil
}
func (c *adaptConn) QueryRow(ctx context.Context, sql string, args ...any) store.Row {
	return c.mock.QueryRow(ctx, sql, args...)
}
func (c *adaptConn) Release() {}

type adaptRows struct{ rows pgx.Rows }

func (r *adaptRows) Next() bool          { return r.rows.Next() }
func (r *adaptRows) Scan(d ...any) error { return r.rows.Scan(d...) }
func (r *adaptRows) Close()              { r.rows.Close() }
func (r *adaptRows) Err() error          { return r.rows.Err() }

func TestSearchCustomer_ReturnsRows(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cols := []string{"id", "name", "email", "phone", "address", "created_at"}
	mock.ExpectQuery(`SELECT id, name, email, phone, address, created_at`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows(cols).
			AddRow(int64(1), "Alice", "a@x.com", (*string)(nil), (*string)(nil), now).
			AddRow(int64(2), "Alicia", "ali@x.com", (*string)(nil), (*string)(nil), now))

	name := "Ali"
	res, err := SearchCustomer(context.Background(), &adaptPool{mock: mock}, SearchCustomerArgs{Name: &name})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Customers) != 2 {
		t.Errorf("rows: %d", len(res.Customers))
	}
	if res.Customers[0].Name != "Alice" {
		t.Errorf("first row: %+v", res.Customers[0])
	}
}

func TestSearchCustomer_EmptyResult(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	cols := []string{"id", "name", "email", "phone", "address", "created_at"}
	mock.ExpectQuery(`SELECT id, name, email, phone, address, created_at`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows(cols))
	name := "Nobody"
	res, err := SearchCustomer(context.Background(), &adaptPool{mock: mock}, SearchCustomerArgs{Name: &name})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Customers) != 0 {
		t.Errorf("expected 0 rows, got %d", len(res.Customers))
	}
}

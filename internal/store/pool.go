package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool is the subset of pgxpool.Pool we depend on (lets tests inject pgxmock).
type Pool interface {
	Acquire(ctx context.Context) (Conn, error)
	Close()
}

// Conn is the subset of *pgxpool.Conn that tools use.
type Conn interface {
	Exec(ctx context.Context, sql string, args ...any) (any, error)
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
	Release()
}

type Row interface {
	Scan(dest ...any) error
}

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close()
	Err() error
}

// pgxPool wraps the concrete pgxpool.Pool into our interface.
type pgxPool struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string, maxConns int, acquireTimeout time.Duration) (Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = int32(maxConns)
	cfg.ConnConfig.RuntimeParams["application_name"] = "agent-sql-mcp"
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &pgxPool{pool: pool}, nil
}

func (p *pgxPool) Acquire(ctx context.Context) (Conn, error) {
	c, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	return &pgxConn{conn: c}, nil
}

func (p *pgxPool) Close() { p.pool.Close() }

type pgxConn struct {
	conn *pgxpool.Conn
}

func (c *pgxConn) Exec(ctx context.Context, sql string, args ...any) (any, error) {
	return c.conn.Exec(ctx, sql, args...)
}

func (c *pgxConn) Query(ctx context.Context, sql string, args ...any) (Rows, error) {
	r, err := c.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &pgxRows{rows: r}, nil
}

func (c *pgxConn) QueryRow(ctx context.Context, sql string, args ...any) Row {
	return c.conn.QueryRow(ctx, sql, args...)
}

func (c *pgxConn) Release() { c.conn.Release() }

type pgxRows struct {
	rows interface {
		Next() bool
		Scan(dest ...any) error
		Close()
		Err() error
	}
}

func (r *pgxRows) Next() bool             { return r.rows.Next() }
func (r *pgxRows) Scan(d ...any) error    { return r.rows.Scan(d...) }
func (r *pgxRows) Close()                 { r.rows.Close() }
func (r *pgxRows) Err() error             { return r.rows.Err() }

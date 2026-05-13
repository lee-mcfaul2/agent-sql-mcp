package tools

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
)

// ErrNotFound is returned when a single-row lookup misses.
var ErrNotFound = errors.New("not found")

type LookupCustomerArgs struct {
	CustomerID int64 `json:"customer_id"`
}

type LookupCustomerResponse struct {
	Customer Customer `json:"customer"`
}

func LookupCustomer(ctx context.Context, p store.Pool, args LookupCustomerArgs) (*LookupCustomerResponse, error) {
	c, err := p.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	row := c.QueryRow(ctx, store.SQLLookupCustomer, args.CustomerID)
	var cust Customer
	if err := row.Scan(&cust.ID, &cust.Name, &cust.Email, &cust.Phone, &cust.Address, &cust.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &LookupCustomerResponse{Customer: cust}, nil
}

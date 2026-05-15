package tools

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/auth"
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

// LookupCustomer runs the lookup_customer tool. claims is used for the
// row-level Atlantis filter — a non-privileged lookup of an atlantis customer
// returns ErrNotFound (the row is filtered out before reaching the handler).
func LookupCustomer(ctx context.Context, p store.Pool, claims auth.UserClaims, args LookupCustomerArgs) (*LookupCustomerResponse, error) {
	c, err := p.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	canSeeAtlantis := claims.HasAll([]string{atlantisReadPerm})

	row := c.QueryRow(ctx, store.SQLLookupCustomer, args.CustomerID, canSeeAtlantis)
	var cust Customer
	if err := row.Scan(&cust.ID, &cust.Name, &cust.Email, &cust.Phone, &cust.Address, &cust.CreatedAt, &cust.Region); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &LookupCustomerResponse{Customer: cust}, nil
}

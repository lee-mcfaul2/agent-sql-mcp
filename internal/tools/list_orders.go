package tools

import (
	"context"
	"time"

	"github.com/lee-mcfaul2/agent-sql-mcp/internal/auth"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
)

// ListOrdersArgs.CustomerID is *int64 (not int64): nil means "across all
// customers" and the SQL's row-level Atlantis filter keeps callers without
// customers:atlantis:read from seeing Atlantis customers' orders. A
// supplied customer_id keeps the existing customer-scoped behaviour.
type ListOrdersArgs struct {
	CustomerID *int64     `json:"customer_id,omitempty"`
	Since      *time.Time `json:"since,omitempty"`
	Limit      *int       `json:"limit,omitempty"`
}

type ListOrdersResponse struct {
	Orders []Order `json:"orders"`
}

func ListOrders(ctx context.Context, p store.Pool, claims auth.UserClaims, args ListOrdersArgs) (*ListOrdersResponse, error) {
	c, err := p.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	limit := 25
	if args.Limit != nil {
		limit = *args.Limit
	}
	canSeeAtlantis := claims.HasAll([]string{atlantisReadPerm})

	rows, err := c.Query(ctx, store.SQLListOrders, args.CustomerID, args.Since, canSeeAtlantis, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := ListOrdersResponse{Orders: []Order{}}
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.Status, &o.TotalCents, &o.Currency, &o.PlacedAt); err != nil {
			return nil, err
		}
		out.Orders = append(out.Orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &out, nil
}

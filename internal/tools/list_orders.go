package tools

import (
	"context"
	"time"

	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
)

type ListOrdersArgs struct {
	CustomerID int64      `json:"customer_id"`
	Since      *time.Time `json:"since,omitempty"`
	Limit      *int       `json:"limit,omitempty"`
}

type ListOrdersResponse struct {
	Orders []Order `json:"orders"`
}

func ListOrders(ctx context.Context, p store.Pool, args ListOrdersArgs) (*ListOrdersResponse, error) {
	c, err := p.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	limit := 25
	if args.Limit != nil {
		limit = *args.Limit
	}

	rows, err := c.Query(ctx, store.SQLListOrders, args.CustomerID, args.Since, limit)
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

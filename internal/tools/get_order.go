package tools

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
)

type GetOrderArgs struct {
	OrderID int64 `json:"order_id"`
}

type GetOrderResponse struct {
	Order OrderWithItems `json:"order"`
}

func GetOrder(ctx context.Context, p store.Pool, args GetOrderArgs) (*GetOrderResponse, error) {
	c, err := p.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	var o Order
	row := c.QueryRow(ctx, store.SQLGetOrder, args.OrderID)
	if err := row.Scan(&o.ID, &o.CustomerID, &o.Status, &o.TotalCents, &o.Currency, &o.PlacedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	rows, err := c.Query(ctx, store.SQLGetOrderItems, args.OrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []OrderItem{}
	for rows.Next() {
		var it OrderItem
		if err := rows.Scan(&it.ID, &it.SKU, &it.Quantity, &it.UnitCents); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &GetOrderResponse{Order: OrderWithItems{Order: o, LineItems: items}}, nil
}

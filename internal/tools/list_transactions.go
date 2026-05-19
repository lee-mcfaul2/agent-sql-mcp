package tools

import (
	"context"
	"time"

	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
)

type ListTransactionsArgs struct {
	CustomerID int64      `json:"customer_id"`
	Since      *time.Time `json:"since,omitempty"`
	Limit      *int       `json:"limit,omitempty"`
}

type ListTransactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
}

func ListTransactions(ctx context.Context, p store.Pool, args ListTransactionsArgs) (*ListTransactionsResponse, error) {
	c, err := p.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	limit := 25
	if args.Limit != nil {
		limit = *args.Limit
	}

	rows, err := c.Query(ctx, store.SQLListTransactions, args.CustomerID, args.Since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := ListTransactionsResponse{Transactions: []Transaction{}}
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(&tx.ID, &tx.CustomerID, &tx.AmountCents, &tx.Kind, &tx.CreatedAt); err != nil {
			return nil, err
		}
		out.Transactions = append(out.Transactions, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &out, nil
}

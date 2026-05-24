package tools

import (
	"context"
	"time"

	"github.com/lee-mcfaul2/agent-sql-mcp/internal/auth"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
)

// ListTransactionsArgs.CustomerID is *int64 (not int64): nil means "across all
// customers" and the SQL's row-level Atlantis filter keeps callers without
// customers:atlantis:read from seeing Atlantis customers' transactions.
type ListTransactionsArgs struct {
	CustomerID *int64     `json:"customer_id,omitempty"`
	Since      *time.Time `json:"since,omitempty"`
	Limit      *int       `json:"limit,omitempty"`
}

type ListTransactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
}

func ListTransactions(ctx context.Context, p store.Pool, claims auth.UserClaims, args ListTransactionsArgs) (*ListTransactionsResponse, error) {
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

	rows, err := c.Query(ctx, store.SQLListTransactions, args.CustomerID, args.Since, canSeeAtlantis, limit)
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

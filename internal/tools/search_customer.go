package tools

import (
	"context"

	"github.com/lee-mcfaul2/agent-sql-mcp/internal/auth"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
)

// atlantisReadPerm is the fine-grained permission that lifts the row-level
// region='atlantis' filter on the customer-tool result set.
const atlantisReadPerm = "customers:atlantis:read"

type SearchCustomerArgs struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
	Phone *string `json:"phone,omitempty"`
}

type SearchCustomerResponse struct {
	Customers []Customer `json:"customers"`
}

// SearchCustomer runs the search_customer tool. claims is used for the
// row-level Atlantis filter — callers lacking customers:atlantis:read get a
// result set with region='atlantis' rows excluded.
func SearchCustomer(ctx context.Context, p store.Pool, claims auth.UserClaims, args SearchCustomerArgs) (*SearchCustomerResponse, error) {
	c, err := p.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	canSeeAtlantis := claims.HasAll([]string{atlantisReadPerm})

	rows, err := c.Query(ctx, store.SQLSearchCustomer, args.Name, args.Email, args.Phone, canSeeAtlantis)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := SearchCustomerResponse{Customers: []Customer{}}
	for rows.Next() {
		var cust Customer
		if err := rows.Scan(&cust.ID, &cust.Name, &cust.Email, &cust.Phone, &cust.Address, &cust.CreatedAt, &cust.Region); err != nil {
			return nil, err
		}
		out.Customers = append(out.Customers, cust)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &out, nil
}

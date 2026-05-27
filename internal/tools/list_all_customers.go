package tools

import (
	"context"

	"github.com/lee-mcfaul2/agent-sql-mcp/internal/auth"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
)

// ListAllCustomersArgs: no filter, just a result cap. Surface targets the
// "browse mode" use case where the agent doesn't yet have any IDs/names to
// search by and just wants the most recent customers. Atlantis row-level
// filter is enforced in the SQL via canSeeAtlantis.
type ListAllCustomersArgs struct {
	Limit *int `json:"limit,omitempty"`
}

type ListAllCustomersResponse struct {
	Customers []Customer `json:"customers"`
}

func ListAllCustomers(ctx context.Context, p store.Pool, claims auth.UserClaims, args ListAllCustomersArgs) (*ListAllCustomersResponse, error) {
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

	rows, err := c.Query(ctx, store.SQLListAllCustomers, canSeeAtlantis, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := ListAllCustomersResponse{Customers: []Customer{}}
	for rows.Next() {
		var c Customer
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Address, &c.CreatedAt, &c.Region); err != nil {
			return nil, err
		}
		out.Customers = append(out.Customers, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &out, nil
}

package tools

import (
	"context"

	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
)

type SearchCustomerArgs struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
	Phone *string `json:"phone,omitempty"`
}

type SearchCustomerResponse struct {
	Customers []Customer `json:"customers"`
}

func SearchCustomer(ctx context.Context, p store.Pool, args SearchCustomerArgs) (*SearchCustomerResponse, error) {
	c, err := p.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	rows, err := c.Query(ctx, store.SQLSearchCustomer, args.Name, args.Email, args.Phone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := SearchCustomerResponse{Customers: []Customer{}}
	for rows.Next() {
		var cust Customer
		if err := rows.Scan(&cust.ID, &cust.Name, &cust.Email, &cust.Phone, &cust.Address, &cust.CreatedAt); err != nil {
			return nil, err
		}
		out.Customers = append(out.Customers, cust)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &out, nil
}

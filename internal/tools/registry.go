package tools

import (
	"context"
	"encoding/json"

	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
)

// ToolFn is the dispatcher-facing signature. Accepts raw (schema-validated) args bytes,
// returns the typed response struct or an error.
type ToolFn func(ctx context.Context, p store.Pool, raw json.RawMessage) (any, error)

// Registry maps tool name to its adapter.
var Registry = map[string]ToolFn{
	"search_customer": adaptSearchCustomer,
	"lookup_customer": adaptLookupCustomer,
	"list_orders":     adaptListOrders,
	"get_order":       adaptGetOrder,
}

func adaptSearchCustomer(ctx context.Context, p store.Pool, raw json.RawMessage) (any, error) {
	var args SearchCustomerArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	return SearchCustomer(ctx, p, args)
}

func adaptLookupCustomer(ctx context.Context, p store.Pool, raw json.RawMessage) (any, error) {
	var args LookupCustomerArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	return LookupCustomer(ctx, p, args)
}

func adaptListOrders(ctx context.Context, p store.Pool, raw json.RawMessage) (any, error) {
	var args ListOrdersArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	return ListOrders(ctx, p, args)
}

func adaptGetOrder(ctx context.Context, p store.Pool, raw json.RawMessage) (any, error) {
	var args GetOrderArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	return GetOrder(ctx, p, args)
}

package tools

import (
	"context"
	"encoding/json"

	"github.com/lee-mcfaul2/agent-sql-mcp/internal/auth"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
)

// ToolFn is the dispatcher-facing signature. Accepts raw (schema-validated) args
// bytes plus the caller's UserClaims (used by per-tool authz logic — currently
// the row-level Atlantis filter on the customer tools). Returns the typed
// response struct or an error.
type ToolFn func(ctx context.Context, p store.Pool, claims auth.UserClaims, raw json.RawMessage) (any, error)

// Registry maps tool name to its adapter.
var Registry = map[string]ToolFn{
	"search_customer":   adaptSearchCustomer,
	"lookup_customer":   adaptLookupCustomer,
	"list_orders":       adaptListOrders,
	"list_transactions": adaptListTransactions,
	"get_order":         adaptGetOrder,
}

func adaptSearchCustomer(ctx context.Context, p store.Pool, claims auth.UserClaims, raw json.RawMessage) (any, error) {
	var args SearchCustomerArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	return SearchCustomer(ctx, p, claims, args)
}

func adaptLookupCustomer(ctx context.Context, p store.Pool, claims auth.UserClaims, raw json.RawMessage) (any, error) {
	var args LookupCustomerArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	return LookupCustomer(ctx, p, claims, args)
}

func adaptListOrders(ctx context.Context, p store.Pool, _ auth.UserClaims, raw json.RawMessage) (any, error) {
	var args ListOrdersArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	return ListOrders(ctx, p, args)
}

func adaptListTransactions(ctx context.Context, p store.Pool, _ auth.UserClaims, raw json.RawMessage) (any, error) {
	var args ListTransactionsArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	return ListTransactions(ctx, p, args)
}

func adaptGetOrder(ctx context.Context, p store.Pool, _ auth.UserClaims, raw json.RawMessage) (any, error) {
	var args GetOrderArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	return GetOrder(ctx, p, args)
}

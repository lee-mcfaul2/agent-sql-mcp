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
	"search_customer":       adaptSearchCustomer,
	"lookup_customer":       adaptLookupCustomer,
	"list_orders":           adaptListOrders,
	"list_transactions":     adaptListTransactions,
	"get_order":             adaptGetOrder,
	// list_all_<table>: thin "no filter, just give me a page" tools so a
	// small LLM doesn't have to compose `list_orders({})`. list_all_orders
	// and list_all_transactions reuse the existing backends with
	// customer_id/since pinned nil; list_all_customers is a new backend
	// (there was no way to enumerate customers without a search query).
	"list_all_customers":    adaptListAllCustomers,
	"list_all_orders":       adaptListAllOrders,
	"list_all_transactions": adaptListAllTransactions,
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

func adaptListOrders(ctx context.Context, p store.Pool, claims auth.UserClaims, raw json.RawMessage) (any, error) {
	var args ListOrdersArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	return ListOrders(ctx, p, claims, args)
}

func adaptListTransactions(ctx context.Context, p store.Pool, claims auth.UserClaims, raw json.RawMessage) (any, error) {
	var args ListTransactionsArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	return ListTransactions(ctx, p, claims, args)
}

func adaptGetOrder(ctx context.Context, p store.Pool, _ auth.UserClaims, raw json.RawMessage) (any, error) {
	var args GetOrderArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	return GetOrder(ctx, p, args)
}

// list_all_<table> adapters. Args schema is a single optional `limit`.
// list_all_orders / list_all_transactions delegate to the existing browse-
// mode handlers with customer_id and since pinned to nil; list_all_customers
// uses its own dedicated handler.

type listAllArgs struct {
	Limit *int `json:"limit,omitempty"`
}

func adaptListAllCustomers(ctx context.Context, p store.Pool, claims auth.UserClaims, raw json.RawMessage) (any, error) {
	var args ListAllCustomersArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	return ListAllCustomers(ctx, p, claims, args)
}

func adaptListAllOrders(ctx context.Context, p store.Pool, claims auth.UserClaims, raw json.RawMessage) (any, error) {
	var args listAllArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	return ListOrders(ctx, p, claims, ListOrdersArgs{Limit: args.Limit})
}

func adaptListAllTransactions(ctx context.Context, p store.Pool, claims auth.UserClaims, raw json.RawMessage) (any, error) {
	var args listAllArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	return ListTransactions(ctx, p, claims, ListTransactionsArgs{Limit: args.Limit})
}

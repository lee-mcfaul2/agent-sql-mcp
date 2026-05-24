package store

// SQL constants. Parameterized; pgx handles placeholders.

const (
	// SQLSearchCustomer: $4 is canSeeAtlantis (bool). When false, the row-level
	// filter excludes region='atlantis' rows; when true, the OR short-circuits
	// and atlantis rows pass through. See ai_security_hybrid_authz: tool-level
	// permission gates the call; this clause is the finer-grained row gate.
	SQLSearchCustomer = `
SELECT id, name, email, phone, address, created_at, region
FROM customers
WHERE
  ($1::text IS NULL OR name  ILIKE '%' || $1 || '%')
  AND ($2::text IS NULL OR email = $2)
  AND ($3::text IS NULL OR phone = $3)
  AND ($4::bool OR region != 'atlantis')
ORDER BY created_at DESC
LIMIT 25
`

	// SQLLookupCustomer: $2 is canSeeAtlantis (bool). Same row-level gate as
	// SQLSearchCustomer; a non-privileged lookup of an atlantis customer
	// returns zero rows (which the handler maps to ErrNotFound).
	SQLLookupCustomer = `
SELECT id, name, email, phone, address, created_at, region
FROM customers
WHERE id = $1
  AND ($2::bool OR region != 'atlantis')
`

	// SQLListOrders: customer_id ($1) is NULLABLE so the agent can browse
	// orders across all customers (necessary for prompts like "how many
	// orders this year" where the agent doesn't yet know any customer_id).
	// Whenever customer_id is omitted, the JOIN+($3) row-level filter is
	// what keeps Atlantis customers' orders hidden from callers without
	// customers:atlantis:read. The customer-scoped path keeps working
	// identically when customer_id is supplied.
	// Params: ($1 customer_id?, $2 since?, $3 canSeeAtlantis, $4 limit).
	SQLListOrders = `
SELECT o.id, o.customer_id, o.status, o.total_cents, o.currency, o.placed_at
FROM orders o
JOIN customers c ON c.id = o.customer_id
WHERE ($1::bigint IS NULL OR o.customer_id = $1)
  AND ($2::timestamptz IS NULL OR o.placed_at >= $2)
  AND ($3::bool OR c.region != 'atlantis')
ORDER BY o.placed_at DESC
LIMIT $4
`

	// SQLListTransactions: same rationale as SQLListOrders -- customer_id
	// is NULLABLE for cross-customer browsing, and the JOIN+($3) filter
	// preserves the Atlantis row-level scope.
	// Params: ($1 customer_id?, $2 since?, $3 canSeeAtlantis, $4 limit).
	SQLListTransactions = `
SELECT t.id, t.customer_id, t.amount_cents, t.kind, t.created_at
FROM transactions t
JOIN customers c ON c.id = t.customer_id
WHERE ($1::bigint IS NULL OR t.customer_id = $1)
  AND ($2::timestamptz IS NULL OR t.created_at >= $2)
  AND ($3::bool OR c.region != 'atlantis')
ORDER BY t.created_at DESC
LIMIT $4
`

	SQLGetOrder = `
SELECT id, customer_id, status, total_cents, currency, placed_at
FROM orders
WHERE id = $1
`

	SQLGetOrderItems = `
SELECT id, sku, quantity, unit_cents
FROM order_items
WHERE order_id = $1
ORDER BY id ASC
`

	SQLInsertAudit = `
INSERT INTO mcp_audit (user_sub, tool, outcome, duration_ms, reason)
VALUES ($1, $2, $3, $4, $5)
`
)

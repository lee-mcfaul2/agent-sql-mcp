package store

// SQL constants. Parameterized; pgx handles placeholders.

const (
	SQLSearchCustomer = `
SELECT id, name, email, phone, address, created_at
FROM customers
WHERE
  ($1::text IS NULL OR name  ILIKE '%' || $1 || '%')
  AND ($2::text IS NULL OR email = $2)
  AND ($3::text IS NULL OR phone = $3)
ORDER BY created_at DESC
LIMIT 25
`

	SQLLookupCustomer = `
SELECT id, name, email, phone, address, created_at
FROM customers
WHERE id = $1
`

	SQLListOrders = `
SELECT id, customer_id, status, total_cents, currency, placed_at
FROM orders
WHERE customer_id = $1
  AND ($2::timestamptz IS NULL OR placed_at >= $2)
ORDER BY placed_at DESC
LIMIT $3
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

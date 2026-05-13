# agent-sql-mcp HTTP API

## Endpoints

| Endpoint | Method | Auth | Purpose |
|---|---|---|---|
| `/handshake` | GET | mesh mTLS (gateway SPIFFE) | Returns `{schema_version, schema_digest}` |
| `/v1/tools/<tool>` | POST | mesh mTLS + `Authorization: Bearer <JWT>` | Invoke a tool |
| `/healthz` | GET | none | Liveness |
| `/readyz` | GET | none | Readiness (pool acquire) |
| `/metrics` | GET | mesh pull | Prometheus exposition |

## Tools

### search_customer

Request: `{ "name"?: string, "email"?: string, "phone"?: string }` (≥1 required)
Response: `{ "customers": [{ id, name, email, phone, address, created_at }, ...] }` (≤25 rows)
Permissions: `customers:read`

### lookup_customer

Request: `{ "customer_id": int }`
Response: `{ "customer": { id, name, email, phone, address, created_at } }` or 404
Permissions: `customers:read`

### list_orders

Request: `{ "customer_id": int, "since"?: ISO8601, "limit"?: 1..100 }`
Response: `{ "orders": [{ id, customer_id, status, total_cents, currency, placed_at }, ...] }`
Permissions: `orders:read`

### get_order

Request: `{ "order_id": int }`
Response: `{ "order": { id, customer_id, status, total_cents, currency, placed_at, line_items: [...] } }` or 404
Permissions: `orders:read`

## Error envelope

```json
{
  "error_type": "PERMISSION_DENIED",
  "retriable": false,
  "message": "missing_permission: orders:read",
  "request_id": "..."
}
```

See spec §7.1 for the full catalog.

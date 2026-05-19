CREATE TABLE IF NOT EXISTS transactions (
    id           BIGSERIAL PRIMARY KEY,
    customer_id  BIGINT NOT NULL REFERENCES customers(id),
    amount_cents BIGINT NOT NULL,
    kind         TEXT NOT NULL CHECK (kind IN ('payment','refund','chargeback')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_transactions_customer ON transactions(customer_id);

-- Deterministic seed: setseed gives reproducible random()s for fixture generation.
SELECT setseed(0.42);

-- ~200 transactions, distributed across the existing 100 customers (ids 1..100).
INSERT INTO transactions (customer_id, amount_cents, kind, created_at)
SELECT
    1 + (g * 7) % 100,
    500 + (g * 13) % 40000,
    (ARRAY['payment','refund','chargeback'])[1 + (g % 3)],
    now() - ((200 - g) || ' hours')::interval
FROM generate_series(1, 200) AS g;

GRANT SELECT ON transactions TO agent_sql_mcp_runtime;

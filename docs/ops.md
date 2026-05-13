# Operations runbook

## Migrations

Migrations are pre-install / pre-upgrade Helm hooks. The Job runs `migrate up` before the Deployment becomes ready.

To re-run manually:
```
kubectl create job --from=cronjob/agent-sql-mcp-migrate manual-migrate-$(date +%s) -n mcp
```

To roll back:
```
kubectl run --rm -it --restart=Never migrate-down -n mcp \
  --image=migrate/migrate:v4.18.1 \
  -- -path=/migrations -database "$SQLMCP_DATABASE_URL" down 1
```

## Audit retention

The `agent-sql-mcp-audit-prune` CronJob deletes `mcp_audit` rows older than `audit.retentionDays` (default 90) every day at 03:00 UTC.

To inspect recent activity:
```sql
SELECT user_sub, count(*), max(ts) AS last_seen
FROM mcp_audit
WHERE ts > now() - interval '1 hour'
GROUP BY user_sub
ORDER BY 2 DESC;
```

## Schema digest mismatch

If the gateway logs `schema_digest_mismatch` for this MCP:
1. Confirm both gateway and MCP point at the same `lib-agent-prompt` release.
2. Run `curl http://agent-sql-mcp.mcp.svc.cluster.local:8443/handshake` from inside the cluster — compare the digest.
3. Rebuild whichever side is stale.

## Common failure modes

| Symptom | Likely cause | Fix |
|---|---|---|
| All `BACKEND_UNAVAILABLE` | Postgres down or pool exhausted | Check pg pod; check `sqlmcp_pg_pool_acquire_seconds` |
| All `JWT_VALIDATION_FAILED` | JWKS cache stale / IdP unreachable | Check `sqlmcp_jwks_refresh_failures_total`; restart pod |
| All `FORBIDDEN_CALLER` | mesh authz misconfigured | Verify gateway SPIFFE matches `SQLMCP_GATEWAY_SPIFFE` |
| Slow queries | Missing index | `EXPLAIN ANALYZE` the slow query; add index in a new migration |

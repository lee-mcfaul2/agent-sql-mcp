# agent-sql-mcp

Customer-support MCP for the AI Agent Security Platform. Four read-only tools (`search_customer`, `lookup_customer`, `list_orders`, `get_order`) backed by a Postgres database. Called by `agent-gateway` over mesh mTLS. Validates the user JWT forwarded by the gateway and runs per-tool permission checks before touching Postgres.

Spec: `ai-security/docs/superpowers/specs/2026-05-13-agent-sql-mcp-design.md` in the umbrella workspace.

## Quickstart

```
make build
make test
make run-local       # boots mcp + postgres via Docker Compose
```

## Layout

- `cmd/sql-mcp/` — entrypoint
- `internal/` — core packages (server, auth, schemas, tools, store, obs)
- `schemas/` — JSON Schema files, embedded at build time
- `migrations/` — Postgres schema + fixture seeds
- `deploy/` — Dockerfile + Helm chart fragment
- `tests/` — unit / integration / conformance
- `docs/` — api, ops, threat-model

## License

Apache 2.0.

# Threat model

## What agent-sql-mcp protects against

1. **Unauthorized callers.** Linkerd `AuthorizationPolicy` rejects any pod that isn't `agent-gateway` (SPIFFE-bound). A belt-and-braces SPIFFE substring check in middleware catches mesh-bypass.
2. **Unauthorized users.** Every `/v1/tools/<tool>` call requires a valid OIDC JWT forwarded by the gateway. Reason-labeled failure metrics surface signature/expiry/audience/issuer issues.
3. **Privilege escalation.** Per-tool permission check: the user's claims must contain every permission required by the tool. Missing → `403 PERMISSION_DENIED` + denial metric.
4. **SQL injection.** All SQL is parameterized via pgx; constants live in `internal/store/queries.go` and a unit test forbids format verbs in SQL.
5. **Backend escalation.** The runtime Postgres role has SELECT-only access to data tables + INSERT-only access to `mcp_audit`. A compromised process cannot mutate customer data, even with correct creds.
6. **Schema drift / fleet-wide silent breakage.** The schema digest is computed at build time and reported via `/handshake`. The gateway shuts down the MCP if the digest doesn't match its catalog.
7. **Response-shape drift.** Each tool response is schema-validated before being returned; a mismatch fires a paging metric.

## What agent-sql-mcp does NOT protect against

1. **A compromised IdP.** Stolen IdP signing keys mean attackers can mint valid JWTs. Defense lives at the IdP layer (KMS-backed keys, rotation, audit).
2. **A compromised gateway.** The gateway is a trusted upstream; if it's compromised, it can send arbitrary detokenized args and arbitrary JWTs. The MCP still runs schema validation and per-tool perms but cannot detect a malicious gateway by design.
3. **A compromised Postgres.** The MCP trusts what Postgres returns. A compromised database can serve fake data.
4. **DoS at the mesh edge.** Rate limiting is the gateway's responsibility; this MCP has no rate-limit logic.
5. **Side-channel attacks on the host.** Standard exclusion.

## Cryptographic primitives

- **JWT validation:** RS256 / ES256 via `lestrrat-go/jwx`, JWKS cached + refreshed periodically.
- **mTLS:** provided by Linkerd between gateway and this MCP. SPIFFE identities issued by SPIRE.
- **TLS to Postgres:** optional, controlled by `?sslmode=` in the DSN. The Helm chart pins `sslmode=require` in production.

## Audit obligations

Every tool call (allow or deny) writes a row to `mcp_audit`. Retention is 90 days by default, enforced by a daily CronJob. The audit row carries raw `sub`; on-call queries it directly via SQL.

## Cross-MCP / fleet considerations

The hybrid-authz design (see `ai_security_hybrid_authz` memory) places per-tool permission code in this repo, not in agent-gateway. Adding a new tool means:
1. New file under `internal/tools/`
2. New schema pair under `schemas/`
3. New entry in `internal/auth/perms.go`
4. New migration if the schema changes

No gateway code change is required to introduce a new tool.

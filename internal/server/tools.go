package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/auth"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/obs"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/tools"
)

func handleTool(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		toolName := chi.URLParam(r, "tool")

		// 1) JWT
		tok, err := auth.ParseBearer(r.Header.Get("Authorization"))
		if err != nil {
			obs.JWTFailuresTotal.WithLabelValues("missing").Inc()
			obs.ToolCallsTotal.WithLabelValues(toolName, "jwt_failed").Inc()
			WriteError(w, r, "JWT_MISSING", err.Error())
			return
		}
		claims, err := deps.JWT.Validate(tok)
		if err != nil {
			obs.ToolCallsTotal.WithLabelValues(toolName, "jwt_failed").Inc()
			WriteError(w, r, "JWT_VALIDATION_FAILED", err.Error())
			return
		}

		// 2) Tool exists?
		if _, ok := tools.Registry[toolName]; !ok {
			obs.AuthzDenialsTotal.WithLabelValues(toolName, "unknown_tool").Inc()
			obs.ToolCallsTotal.WithLabelValues(toolName, "not_found").Inc()
			WriteError(w, r, "UNKNOWN_TOOL", toolName)
			return
		}

		// 3) Permission check
		required, _ := auth.RequiredFor(toolName)
		// REVERT-BEFORE-RELEASE: unsafe verbose debug for permission check
		slog.Default().Info("tool.perm_check.unsafe_debug",
			"tool", toolName,
			"claims_sub", claims.Sub,
			"claims_permissions", claims.Permissions,
			"claims_groups", claims.Groups,
			"required", required,
			"missing", claims.Missing(required),
		)
		if missing := claims.Missing(required); len(missing) > 0 {
			obs.AuthzDenialsTotal.WithLabelValues(toolName, "missing_permission").Inc()
			obs.ToolCallsTotal.WithLabelValues(toolName, "permission_denied").Inc()
			writeAuditAsync(deps, claims.Sub, toolName, "permission_denied", time.Since(start), fmt.Sprintf("missing_permission:%s", missing[0]))
			WriteError(w, r, "PERMISSION_DENIED", fmt.Sprintf("missing_permission:%s", missing[0]))
			return
		}

		// 4) Decode body
		var rawArgs json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&rawArgs); err != nil {
			obs.ToolCallsTotal.WithLabelValues(toolName, "validation").Inc()
			WriteError(w, r, "BAD_REQUEST", err.Error())
			return
		}

		// 5) Schema-validate request
		anyArgs, err := jsonToAny(rawArgs)
		if err != nil {
			obs.ToolCallsTotal.WithLabelValues(toolName, "validation").Inc()
			WriteError(w, r, "BAD_REQUEST", err.Error())
			return
		}
		if err := deps.Validators.ValidateRequest(toolName, anyArgs); err != nil {
			obs.SchemaFailuresTotal.WithLabelValues(toolName, "request").Inc()
			obs.ToolCallsTotal.WithLabelValues(toolName, "schema_failed").Inc()
			WriteError(w, r, "SCHEMA_VALIDATION_FAILED", err.Error())
			return
		}

		// 6-8) Execute with deadline
		qctx, cancel := context.WithTimeout(r.Context(), time.Duration(deps.QueryTimeout())*time.Second)
		defer cancel()

		result, runErr := tools.Registry[toolName](qctx, deps.Pool, claims, rawArgs)

		if runErr != nil {
			et, reason := classifyBackendError(runErr)
			obs.ToolCallsTotal.WithLabelValues(toolName, et).Inc()
			obs.ToolDuration.WithLabelValues(toolName, et).Observe(time.Since(start).Seconds())
			writeAuditAsync(deps, claims.Sub, toolName, et, time.Since(start), reason)
			switch et {
			case "not_found":
				WriteError(w, r, "NOT_FOUND", reason)
			case "timeout":
				WriteError(w, r, "QUERY_TIMEOUT", reason)
			case "unavailable":
				WriteError(w, r, "BACKEND_UNAVAILABLE", reason)
			default:
				WriteError(w, r, "BACKEND_ERROR", reason)
			}
			return
		}

		// 9) Schema-validate response
		respBytes, err := json.Marshal(result)
		if err != nil {
			obs.ToolCallsTotal.WithLabelValues(toolName, "backend_error").Inc()
			WriteError(w, r, "INTERNAL_ERROR", err.Error())
			return
		}
		anyResp, _ := jsonToAny(respBytes)
		if err := deps.Validators.ValidateResponse(toolName, anyResp); err != nil {
			obs.SchemaFailuresTotal.WithLabelValues(toolName, "response").Inc()
			obs.ToolCallsTotal.WithLabelValues(toolName, "schema_failed").Inc()
			WriteError(w, r, "INTERNAL_ERROR", "response schema drift")
			return
		}

		// 10) Audit (async — failure does not affect response)
		writeAuditAsync(deps, claims.Sub, toolName, "ok", time.Since(start), "")

		obs.ToolCallsTotal.WithLabelValues(toolName, "ok").Inc()
		obs.ToolDuration.WithLabelValues(toolName, "ok").Observe(time.Since(start).Seconds())

		// 11) Return
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(respBytes)
	}
}

// jsonToAny re-decodes JSON bytes into the `any` tree the validator wants.
func jsonToAny(b []byte) (any, error) {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// classifyBackendError maps a tool-layer error to the canonical envelope error_type + reason.
func classifyBackendError(err error) (errorType, reason string) {
	switch {
	case errors.Is(err, tools.ErrNotFound):
		return "not_found", err.Error()
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout", "query exceeded deadline"
	case strings.Contains(err.Error(), "connection"):
		return "unavailable", err.Error()
	default:
		return "backend_error", err.Error()
	}
}

// writeAuditAsync is best-effort.
func writeAuditAsync(deps Deps, sub, tool, outcome string, dur time.Duration, reason string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := store.WriteAudit(ctx, deps.Pool, store.AuditEntry{
			UserSub:    sub,
			Tool:       tool,
			Outcome:    outcome,
			DurationMs: int(dur.Milliseconds()),
			Reason:     reason,
		})
		if err != nil {
			obs.AuditWritesTotal.WithLabelValues("error").Inc()
			deps.Log.Warn("audit.write_failed", "err", err.Error())
			return
		}
		obs.AuditWritesTotal.WithLabelValues("ok").Inc()
	}()
}

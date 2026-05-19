package auth

// demoPermsBySub is the static demo permission policy keyed by the OIDC `sub`
// claim. It exists ONLY for the demo deployment, where Dex's static/local
// connector cannot emit per-user `permissions`/`groups` claims — it only issues
// a stable per-user `sub` equal to the configured userID. Permissions must
// therefore be derived from `sub`.
//
// This map is consulted as the LAST fallback when building UserClaims.
// Precedence (see derivePermissions): an explicit `permissions` claim wins;
// failing that, a `groups`-derived set; and only if neither is present do we
// fall back to this sub map. A real IdP that emits permissions/groups is
// therefore unaffected.
var demoPermsBySub = map[string][]string{
	// alice — full access (incl. row-level Atlantis customers).
	"00000000-0000-0000-0000-000000000001": {
		"customers:read",
		"customers:atlantis:read",
		"orders:read",
		"transactions:read",
	},
	// bob — customers + orders only (NO transactions, NO atlantis).
	"00000000-0000-0000-0000-000000000002": {
		"customers:read",
		"orders:read",
	},
	// carol — knowledge-base-only; no agent-sql-mcp tool perms, so every
	// SQL tool is correctly denied for her.
	"00000000-0000-0000-0000-000000000003": {},
}

// permsFromGroups maps a derived-from-`groups` policy. The demo IdP cannot
// emit groups either, so this is currently empty; it is kept as an explicit
// extension point so the precedence chain (permissions → groups → sub) is
// visible and a future real IdP that emits groups can be wired here without
// touching the precedence logic.
func permsFromGroups(groups []string) []string {
	if len(groups) == 0 {
		return nil
	}
	// No demo group→perm mappings defined. A real IdP integration would
	// translate group memberships into permission scopes here.
	return nil
}

// derivePermissions resolves the effective permission set for a token using a
// fixed precedence so a future real IdP keeps working:
//
//  1. explicit `permissions` claim, if non-empty
//  2. else permissions derived from the `groups` claim, if any
//  3. else the static demo sub map (returns nil for unknown sub)
func derivePermissions(sub string, permsClaim, groupsClaim []string) []string {
	if len(permsClaim) > 0 {
		return permsClaim
	}
	if g := permsFromGroups(groupsClaim); len(g) > 0 {
		return g
	}
	if p, ok := demoPermsBySub[sub]; ok {
		// Return a copy so callers can't mutate the policy table.
		out := make([]string, len(p))
		copy(out, p)
		return out
	}
	return nil
}

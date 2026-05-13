package auth

// UserClaims is the subset of OIDC claims this MCP cares about.
type UserClaims struct {
	Sub         string
	Permissions []string
	Groups      []string
}

// HasAll returns true if every required perm is in the user's set.
func (c UserClaims) HasAll(required []string) bool {
	have := map[string]struct{}{}
	for _, p := range c.Permissions {
		have[p] = struct{}{}
	}
	for _, r := range required {
		if _, ok := have[r]; !ok {
			return false
		}
	}
	return true
}

// Missing returns the permissions the user is missing (empty if all present).
func (c UserClaims) Missing(required []string) []string {
	have := map[string]struct{}{}
	for _, p := range c.Permissions {
		have[p] = struct{}{}
	}
	var missing []string
	for _, r := range required {
		if _, ok := have[r]; !ok {
			missing = append(missing, r)
		}
	}
	return missing
}

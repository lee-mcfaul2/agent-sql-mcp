package auth

import "fmt"

// ToolPerms maps tool name -> required permissions.
// Single source of truth for what each tool needs.
var ToolPerms = map[string][]string{
	"search_customer": {"customers:read"},
	"lookup_customer": {"customers:read"},
	"list_orders":     {"orders:read"},
	"get_order":       {"orders:read"},
}

// RequiredFor returns the permissions required for tool, or an error if the tool is unknown.
func RequiredFor(tool string) ([]string, error) {
	p, ok := ToolPerms[tool]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", tool)
	}
	return p, nil
}

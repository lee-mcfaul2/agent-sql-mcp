package auth

import (
	"context"
	"testing"
	"time"
)

const (
	subAlice = "00000000-0000-0000-0000-000000000001"
	subBob   = "00000000-0000-0000-0000-000000000002"
	subCarol = "00000000-0000-0000-0000-000000000003"
)

// validatorFor spins up a test IdP + validator (reuses helpers from oidc_test.go).
func validatorFor(t *testing.T) (*Validator, func(claims map[string]any) string) {
	t.Helper()
	srv, priv := newTestIdP(t)
	v, err := NewValidator(context.Background(), "http://idp.test", "agent-sql-mcp", srv.URL, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return v, func(claims map[string]any) string { return mint(t, priv, claims) }
}

func base(sub string) map[string]any {
	return map[string]any{"iss": "http://idp.test", "aud": "agent-sql-mcp", "sub": sub}
}

func TestDemoPerms_AliceFullAccess(t *testing.T) {
	v, sign := validatorFor(t)
	claims, err := v.Validate(sign(base(subAlice)))
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	// Every tool's perms, plus the row-level Atlantis gate.
	for tool, req := range ToolPerms {
		if !claims.HasAll(req) {
			t.Errorf("alice should satisfy %s perms %v", tool, req)
		}
	}
	if !claims.HasAll([]string{"customers:atlantis:read"}) {
		t.Error("alice should have customers:atlantis:read")
	}
}

func TestDemoPerms_BobScoped(t *testing.T) {
	v, sign := validatorFor(t)
	claims, err := v.Validate(sign(base(subBob)))
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	// Allowed: lookup_customer (customers:read), list_orders (orders:read).
	if !claims.HasAll(ToolPerms["lookup_customer"]) {
		t.Error("bob should be able to lookup_customer")
	}
	if !claims.HasAll(ToolPerms["list_orders"]) {
		t.Error("bob should be able to list_orders")
	}
	// Denied: list_transactions (transactions:read) and the atlantis gate.
	if claims.HasAll(ToolPerms["list_transactions"]) {
		t.Error("bob must NOT be able to list_transactions")
	}
	if claims.HasAll([]string{"customers:atlantis:read"}) {
		t.Error("bob must NOT have customers:atlantis:read")
	}
}

func TestDemoPerms_CarolDeniedAll(t *testing.T) {
	v, sign := validatorFor(t)
	claims, err := v.Validate(sign(base(subCarol)))
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if len(claims.Permissions) != 0 {
		t.Errorf("carol should have no perms, got %v", claims.Permissions)
	}
	for tool, req := range ToolPerms {
		if claims.HasAll(req) {
			t.Errorf("carol must NOT satisfy %s perms %v", tool, req)
		}
	}
}

func TestDemoPerms_UnknownSubDeniedAll(t *testing.T) {
	v, sign := validatorFor(t)
	claims, err := v.Validate(sign(base("00000000-0000-0000-0000-00000000ffff")))
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if len(claims.Permissions) != 0 {
		t.Errorf("unknown sub should have no perms, got %v", claims.Permissions)
	}
}

func TestDemoPerms_ExplicitClaimOverridesSubMap(t *testing.T) {
	v, sign := validatorFor(t)
	// carol's sub maps to []; an explicit permissions claim must still win
	// so a real IdP that emits the claim keeps working.
	c := base(subCarol)
	c["permissions"] = []string{"orders:read"}
	claims, err := v.Validate(sign(c))
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if !claims.HasAll([]string{"orders:read"}) {
		t.Error("explicit permissions claim should override the demo sub map")
	}
	// And it must NOT be augmented by the (fuller) demo policy for some other sub.
	if claims.HasAll([]string{"customers:read"}) {
		t.Error("explicit claim should be authoritative, not merged with sub map")
	}
}

func TestDerivePermissions_PrecedenceUnit(t *testing.T) {
	// 1. explicit permissions claim wins outright.
	if got := derivePermissions(subAlice, []string{"orders:read"}, []string{"grp"}); len(got) != 1 || got[0] != "orders:read" {
		t.Errorf("explicit perms should win, got %v", got)
	}
	// 2. no explicit perms, groups currently derive nothing → sub map.
	got := derivePermissions(subBob, nil, []string{"some-group"})
	if !contains(got, "customers:read") || !contains(got, "orders:read") || contains(got, "transactions:read") {
		t.Errorf("bob sub-map fallback wrong: %v", got)
	}
	// 3. unknown sub, no claims → nil.
	if got := derivePermissions("nope", nil, nil); got != nil {
		t.Errorf("unknown sub should yield nil, got %v", got)
	}
	// 4. returned slice is a copy (mutation must not corrupt policy table).
	g := derivePermissions(subAlice, nil, nil)
	g[0] = "tampered"
	if demoPermsBySub[subAlice][0] == "tampered" {
		t.Error("derivePermissions must not return the backing slice")
	}
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

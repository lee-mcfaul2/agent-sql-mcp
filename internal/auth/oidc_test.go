package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func newTestIdP(t *testing.T) (*httptest.Server, jwk.Key) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privKey, err := jwk.FromRaw(priv)
	if err != nil {
		t.Fatal(err)
	}
	if err := privKey.Set(jwk.KeyIDKey, "test-kid"); err != nil {
		t.Fatal(err)
	}
	pubKey, err := privKey.PublicKey()
	if err != nil {
		t.Fatal(err)
	}
	if err := pubKey.Set(jwk.AlgorithmKey, jwa.RS256); err != nil {
		t.Fatal(err)
	}
	set := jwk.NewSet()
	if err := set.AddKey(pubKey); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(set)
	}))
	t.Cleanup(srv.Close)
	return srv, privKey
}

func mint(t *testing.T, key jwk.Key, claims map[string]any) string {
	t.Helper()
	tok := jwt.New()
	for k, v := range claims {
		_ = tok.Set(k, v)
	}
	if _, ok := claims["iat"]; !ok {
		_ = tok.Set("iat", time.Now().Unix())
	}
	if _, ok := claims["exp"]; !ok {
		_ = tok.Set("exp", time.Now().Add(time.Hour).Unix())
	}
	b, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, key))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestValidate_OK(t *testing.T) {
	srv, priv := newTestIdP(t)
	v, err := NewValidator(context.Background(), "http://idp.test", "agent-sql-mcp", srv.URL, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	tok := mint(t, priv, map[string]any{
		"iss":         "http://idp.test",
		"aud":         "agent-sql-mcp",
		"sub":         "alice",
		"permissions": []string{"customers:read"},
	})
	claims, err := v.Validate(tok)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Sub != "alice" {
		t.Errorf("sub: %q", claims.Sub)
	}
	if !claims.HasAll([]string{"customers:read"}) {
		t.Error("perms not parsed")
	}
}

func TestValidate_WrongAudience(t *testing.T) {
	srv, priv := newTestIdP(t)
	v, _ := NewValidator(context.Background(), "http://idp.test", "agent-sql-mcp", srv.URL, time.Minute)
	tok := mint(t, priv, map[string]any{
		"iss": "http://idp.test", "aud": "elsewhere", "sub": "alice",
	})
	_, err := v.Validate(tok)
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if ve.Reason != "audience" {
		t.Errorf("reason: %q", ve.Reason)
	}
}

func TestValidate_Expired(t *testing.T) {
	srv, priv := newTestIdP(t)
	v, _ := NewValidator(context.Background(), "http://idp.test", "agent-sql-mcp", srv.URL, time.Minute)
	tok := mint(t, priv, map[string]any{
		"iss": "http://idp.test", "aud": "agent-sql-mcp", "sub": "alice",
		"exp": time.Now().Add(-time.Minute).Unix(),
		"iat": time.Now().Add(-2 * time.Minute).Unix(),
	})
	_, err := v.Validate(tok)
	ve := err.(*ValidationError)
	if ve.Reason != "expired" {
		t.Errorf("reason: %q", ve.Reason)
	}
}

func TestValidate_MissingSub(t *testing.T) {
	srv, priv := newTestIdP(t)
	v, _ := NewValidator(context.Background(), "http://idp.test", "agent-sql-mcp", srv.URL, time.Minute)
	tok := mint(t, priv, map[string]any{
		"iss": "http://idp.test", "aud": "agent-sql-mcp",
	})
	_, err := v.Validate(tok)
	ve := err.(*ValidationError)
	if ve.Reason != "missing_claim" {
		t.Errorf("reason: %q", ve.Reason)
	}
}

func TestParseBearer(t *testing.T) {
	if _, err := ParseBearer("Token abc"); err == nil {
		t.Error("expected error for non-Bearer")
	}
	tok, err := ParseBearer("Bearer abc")
	if err != nil || tok != "abc" {
		t.Errorf("got %q err=%v", tok, err)
	}
}

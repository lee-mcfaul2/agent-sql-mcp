package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/lee-mcfaul2/agent-sql-mcp/internal/obs"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// DiscoverJWKS resolves the JWKS endpoint via OIDC discovery: it fetches
// <issuer>/.well-known/openid-configuration and returns its jwks_uri field.
// This matches agent-gateway's behaviour and works with any spec-compliant
// IdP (e.g. Dex serves keys at <issuer>/keys, not <issuer>/.well-known/jwks.json).
func DiscoverJWKS(ctx context.Context, issuer string) (string, error) {
	discoveryURL := strings.TrimRight(issuer, "/") + "/.well-known/openid-configuration"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return "", fmt.Errorf("build discovery request: %w", err)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch %q: %w", discoveryURL, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("discovery %q: non-200 response code %q", discoveryURL, resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read discovery body: %w", err)
	}
	var doc struct {
		JWKSURI string `json:"jwks_uri"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return "", fmt.Errorf("parse discovery doc: %w", err)
	}
	if doc.JWKSURI == "" {
		return "", fmt.Errorf("discovery %q: missing jwks_uri", discoveryURL)
	}
	return doc.JWKSURI, nil
}

type Validator struct {
	issuer   string
	audience string
	keyset   jwk.Set
}

// NewValidator builds a validator backed by a refreshing JWKS cache.
func NewValidator(ctx context.Context, issuer, audience, jwksURL string, refresh time.Duration) (*Validator, error) {
	cache := jwk.NewCache(ctx)
	if err := cache.Register(jwksURL, jwk.WithMinRefreshInterval(refresh)); err != nil {
		return nil, err
	}
	if _, err := cache.Refresh(ctx, jwksURL); err != nil {
		return nil, fmt.Errorf("initial JWKS fetch: %w", err)
	}
	return &Validator{
		issuer:   issuer,
		audience: audience,
		keyset:   jwk.NewCachedSet(cache, jwksURL),
	}, nil
}

// ParseBearer extracts the token from an Authorization header value.
func ParseBearer(h string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return "", errors.New("missing or malformed Bearer prefix")
	}
	return strings.TrimSpace(h[len(prefix):]), nil
}

// Validate parses + verifies the token and returns the trimmed claims.
func (v *Validator) Validate(token string) (UserClaims, error) {
	t, err := jwt.Parse([]byte(token),
		jwt.WithKeySet(v.keyset),
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithValidate(true),
	)
	if err != nil {
		reason := classify(err)
		obs.JWTFailuresTotal.WithLabelValues(reason).Inc()
		return UserClaims{}, &ValidationError{Reason: reason, Err: err}
	}
	sub, _ := t.Get("sub")
	if sub == nil || sub.(string) == "" {
		obs.JWTFailuresTotal.WithLabelValues("missing_claim").Inc()
		return UserClaims{}, &ValidationError{Reason: "missing_claim", Err: errors.New("sub claim missing")}
	}
	perms := claimsStringSlice(t, "permissions")
	groups := claimsStringSlice(t, "groups")
	return UserClaims{Sub: sub.(string), Permissions: perms, Groups: groups}, nil
}

func classify(err error) string {
	s := err.Error()
	switch {
	case strings.Contains(s, "exp"):
		return "expired"
	case strings.Contains(s, "aud"):
		return "audience"
	case strings.Contains(s, "iss"):
		return "issuer"
	case strings.Contains(s, "signature"):
		return "signature"
	case strings.Contains(s, "key"):
		return "signature"
	}
	return "format"
}

func claimsStringSlice(t jwt.Token, key string) []string {
	raw, ok := t.Get(key)
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case []any:
		out := make([]string, 0, len(v))
		for _, it := range v {
			if s, ok := it.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return v
	default:
		return nil
	}
}

// ValidationError carries the failure reason for metrics + envelope mapping.
type ValidationError struct {
	Reason string // signature | expired | audience | issuer | format | missing_claim
	Err    error
}

func (e *ValidationError) Error() string { return fmt.Sprintf("jwt %s: %v", e.Reason, e.Err) }

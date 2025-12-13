package middleware

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ClerkJWTValidator は Clerk が発行する JWT を JWKS を使って検証するためのヘルパーです。
// 外部依存（JWTライブラリ）を追加せず、RS256 のみを最小実装でサポートします。
type ClerkJWTValidator struct {
	jwksURL  string
	issuer   string
	audience string

	client *http.Client
	cache  *jwksCache
}

// NewClerkJWTValidator は検証器を生成します。
// jwksURL は必須です。issuer/audience は空なら検証しません。
func NewClerkJWTValidator(jwksURL, issuer, audience string) (*ClerkJWTValidator, error) {
	jwksURL = strings.TrimSpace(jwksURL)
	if jwksURL == "" {
		return nil, errors.New("CLERK_JWKS_URL is required")
	}

	v := &ClerkJWTValidator{
		jwksURL:  jwksURL,
		issuer:   strings.TrimSpace(issuer),
		audience: strings.TrimSpace(audience),
		client:   &http.Client{Timeout: 5 * time.Second},
	}
	v.cache = newJWKSCache(v.client, jwksURL, 15*time.Minute)
	return v, nil
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

// Verify は JWT の署名/期限/（任意でiss/aud）を検証し、payload(claims)を返します。
func (v *ClerkJWTValidator) Verify(ctx context.Context, token string) (map[string]any, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("empty token")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("invalid token header encoding")
	}
	var h jwtHeader
	if err := json.Unmarshal(headerJSON, &h); err != nil {
		return nil, errors.New("invalid token header json")
	}
	if h.Alg != "RS256" {
		return nil, fmt.Errorf("unsupported jwt alg: %s", h.Alg)
	}
	if strings.TrimSpace(h.Kid) == "" {
		return nil, errors.New("missing kid in jwt header")
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid token payload encoding")
	}
	claims := map[string]any{}
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, errors.New("invalid token payload json")
	}

	// exp/nbf の検証（存在する場合）
	now := time.Now().Unix()
	if exp, ok := getNumericClaim(claims, "exp"); ok {
		if now >= exp {
			return nil, errors.New("token expired")
		}
	}
	if nbf, ok := getNumericClaim(claims, "nbf"); ok {
		if now < nbf {
			return nil, errors.New("token not yet valid")
		}
	}

	// iss/aud の検証（設定されている場合のみ）
	if v.issuer != "" {
		if iss, _ := claims["iss"].(string); iss != v.issuer {
			return nil, errors.New("invalid issuer")
		}
	}
	if v.audience != "" {
		if !audMatches(claims["aud"], v.audience) {
			return nil, errors.New("invalid audience")
		}
	}

	// 署名検証
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("invalid token signature encoding")
	}
	signingInput := parts[0] + "." + parts[1]
	sum := sha256.Sum256([]byte(signingInput))

	pubAny, err := v.cache.getKey(ctx, h.Kid)
	if err != nil {
		return nil, fmt.Errorf("failed to get jwk: %w", err)
	}
	pub, ok := pubAny.(*rsa.PublicKey)
	if !ok || pub == nil {
		return nil, errors.New("invalid jwk type")
	}
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, sum[:], sig); err != nil {
		return nil, errors.New("invalid token signature")
	}

	return claims, nil
}

func getNumericClaim(claims map[string]any, key string) (int64, bool) {
	v, ok := claims[key]
	if !ok || v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		return int64(t), true
	case json.Number:
		n, err := t.Int64()
		if err != nil {
			return 0, false
		}
		return n, true
	case int64:
		return t, true
	case int:
		return int64(t), true
	default:
		return 0, false
	}
}

func audMatches(aud any, expected string) bool {
	switch t := aud.(type) {
	case string:
		return t == expected
	case []any:
		for _, v := range t {
			if s, ok := v.(string); ok && s == expected {
				return true
			}
		}
		return false
	case []string:
		for _, s := range t {
			if s == expected {
				return true
			}
		}
		return false
	default:
		return false
	}
}

type jwksCache struct {
	client *http.Client
	url    string
	ttl    time.Duration

	mu        sync.RWMutex
	fetchedAt time.Time
	keys      map[string]*rsa.PublicKey
}

func newJWKSCache(client *http.Client, url string, ttl time.Duration) *jwksCache {
	return &jwksCache{
		client: client,
		url:    url,
		ttl:    ttl,
		keys:   map[string]*rsa.PublicKey{},
	}
}

type jwksResponse struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (c *jwksCache) getKey(ctx context.Context, kid string) (any, error) {
	// まずキャッシュヒットを狙う
	c.mu.RLock()
	key, ok := c.keys[kid]
	fetchedAt := c.fetchedAt
	c.mu.RUnlock()

	if ok && key != nil && time.Since(fetchedAt) < c.ttl {
		return key, nil
	}

	// miss or stale なら refresh
	if err := c.refresh(ctx); err != nil {
		// refresh が失敗しても、古いキャッシュに目的のkidがあれば使う（ベストエフォート）
		c.mu.RLock()
		key2, ok2 := c.keys[kid]
		c.mu.RUnlock()
		if ok2 && key2 != nil {
			return key2, nil
		}
		return nil, err
	}

	c.mu.RLock()
	key, ok = c.keys[kid]
	c.mu.RUnlock()
	if !ok || key == nil {
		return nil, errors.New("kid not found in jwks")
	}
	return key, nil
}

func (c *jwksCache) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("jwks fetch failed: status=%d", resp.StatusCode)
	}

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	next := map[string]*rsa.PublicKey{}
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" || strings.TrimSpace(k.Kid) == "" || k.N == "" || k.E == "" {
			continue
		}
		pub, err := rsaPublicKeyFromJWK(k.N, k.E)
		if err != nil {
			continue
		}
		next[k.Kid] = pub
	}

	if len(next) == 0 {
		return errors.New("jwks contains no usable keys")
	}

	c.mu.Lock()
	c.keys = next
	c.fetchedAt = time.Now()
	c.mu.Unlock()

	return nil
}

func rsaPublicKeyFromJWK(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, err
	}
	if len(eBytes) == 0 {
		return nil, errors.New("empty exponent")
	}

	n := new(big.Int).SetBytes(nBytes)
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	if e == 0 {
		return nil, errors.New("invalid exponent")
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}

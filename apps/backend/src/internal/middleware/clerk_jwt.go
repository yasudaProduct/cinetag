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

// Clerk が発行する JWT を JWKS を使って検証するためのヘルパー。
// 外部依存（JWTライブラリ）を追加せず、RS256 のみを最小実装でサポートする。
type ClerkJWTValidator struct {
	jwksURL  string
	issuer   string
	audience string

	client *http.Client
	cache  *jwksCache
}

// Clerk JWT 検証器を生成する。
// - jwksURL は必須。
// - issuer/audience は空の場合は検証しない。
func NewClerkJWTValidator(jwksURL, issuer, audience string) (*ClerkJWTValidator, error) {
	jwksURL = strings.TrimSpace(jwksURL)
	if jwksURL == "" {
		return nil, errors.New("CLERK_JWKS_URL is required")
	}

	// Clerk JWT 検証器を生成。
	v := &ClerkJWTValidator{
		jwksURL:  jwksURL,
		issuer:   strings.TrimSpace(issuer),
		audience: strings.TrimSpace(audience),
		client:   &http.Client{Timeout: 5 * time.Second},
	}
	v.cache = newJWKSCache(v.client, jwksURL, 15*time.Minute)
	return v, nil
}

// JWT ヘッダーの構造。
type jwtHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

// JWT の署名/期限/（任意でiss/aud）を検証し、payload(claims)を返す。
func (v *ClerkJWTValidator) Verify(ctx context.Context, token string) (map[string]any, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("empty token")
	}

	// JWT をパース。
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	// JWT ヘッダーをデコード。
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("invalid token header encoding")
	}

	// JWT ヘッダーをパース。
	var h jwtHeader
	if err := json.Unmarshal(headerJSON, &h); err != nil {
		return nil, errors.New("invalid token header json")
	}

	// JWT アルゴリズムが RS256 でない場合はエラー。
	if h.Alg != "RS256" {
		return nil, fmt.Errorf("unsupported jwt alg: %s", h.Alg)
	}

	// JWT ヘッダーに kid がない場合はエラー。
	if strings.TrimSpace(h.Kid) == "" {
		return nil, errors.New("missing kid in jwt header")
	}

	// JWT ペイロードをデコード。
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid token payload encoding")
	}

	// JWT ペイロードをパース。
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

	// JWT 署名をデコード。
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("invalid token signature encoding")
	}

	// JWT 署名を検証。
	signingInput := parts[0] + "." + parts[1]
	sum := sha256.Sum256([]byte(signingInput))

	// JWK を取得。
	pubAny, err := v.cache.getKey(ctx, h.Kid)
	if err != nil {
		return nil, fmt.Errorf("failed to get jwk: %w", err)
	}

	// JWK をパース。
	pub, ok := pubAny.(*rsa.PublicKey)
	if !ok || pub == nil {
		return nil, errors.New("invalid jwk type")
	}

	// JWT 署名を検証。
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, sum[:], sig); err != nil {
		return nil, errors.New("invalid token signature")
	}

	return claims, nil
}

// JWT ペイロードの数値型のクレームを取得する。
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

// JWT ペイロードの aud が expected と一致するか確認する。
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

// JWKS キャッシュの構造。
type jwksCache struct {
	client *http.Client
	url    string
	ttl    time.Duration

	mu        sync.RWMutex
	fetchedAt time.Time
	keys      map[string]*rsa.PublicKey
}

// JWKS キャッシュを生成する。
func newJWKSCache(client *http.Client, url string, ttl time.Duration) *jwksCache {
	return &jwksCache{
		client: client,
		url:    url,
		ttl:    ttl,
		keys:   map[string]*rsa.PublicKey{},
	}
}

// JWKS レスポンスの構造。
type jwksResponse struct {
	Keys []jwk `json:"keys"`
}

// JWK の構造。
type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWK を取得する。
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

// JWKS をリフレッシュする。
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

// JWK を RSA 公開鍵に変換する。
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

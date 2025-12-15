package middleware

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testJWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type testJWKS struct {
	Keys []testJWK `json:"keys"`
}

func mustNewRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("RSA鍵の生成に失敗: %v", err)
	}
	return key
}

func mustBase64URLJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("JSONの生成に失敗: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func jwkFromPublicKey(kid string, pub *rsa.PublicKey) testJWK {
	// e は 65537 前提（Go の RSA生成は通常 65537）。
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString([]byte{0x01, 0x00, 0x01})
	return testJWK{
		Kty: "RSA",
		Kid: kid,
		Use: "sig",
		Alg: "RS256",
		N:   n,
		E:   e,
	}
}

func newJWKSServer(t *testing.T, jwks testJWKS) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	}))
}

func mustSignRS256JWT(t *testing.T, kid string, claims map[string]any, priv *rsa.PrivateKey) string {
	t.Helper()

	header := map[string]any{
		"alg": "RS256",
		"kid": kid,
		"typ": "JWT",
	}

	encodedHeader := mustBase64URLJSON(t, header)
	encodedPayload := mustBase64URLJSON(t, claims)
	signingInput := encodedHeader + "." + encodedPayload

	sum := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, sum[:])
	if err != nil {
		t.Fatalf("JWT署名に失敗: %v", err)
	}

	encodedSig := base64.RawURLEncoding.EncodeToString(sig)
	return signingInput + "." + encodedSig
}

func mustBuildJWTWithHeader(t *testing.T, header map[string]any, claims map[string]any, signatureB64URL string) string {
	t.Helper()
	encodedHeader := mustBase64URLJSON(t, header)
	encodedPayload := mustBase64URLJSON(t, claims)
	return encodedHeader + "." + encodedPayload + "." + signatureB64URL
}

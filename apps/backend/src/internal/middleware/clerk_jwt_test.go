package middleware

import (
	"context"
	"testing"
	"time"
)

func TestClerkJWTValidator_Verify(t *testing.T) {
	t.Parallel()

	t.Run("成功: RS256署名が検証でき、claims が返る", func(t *testing.T) {
		t.Parallel()

		kid := "kid1"
		priv := mustNewRSAKey(t)
		jwks := testJWKS{Keys: []testJWK{jwkFromPublicKey(kid, &priv.PublicKey)}}
		srv := newJWKSServer(t, jwks)
		t.Cleanup(srv.Close)

		v, err := NewClerkJWTValidator(srv.URL, "", "")
		if err != nil {
			t.Fatalf("validator生成に失敗: %v", err)
		}

		claims := map[string]any{
			"sub":   "user_123",
			"email": "a@example.com",
			"exp":   time.Now().Add(10 * time.Minute).Unix(),
			"nbf":   time.Now().Add(-1 * time.Minute).Unix(),
		}
		token := mustSignRS256JWT(t, kid, claims, priv)

		got, err := v.Verify(context.Background(), token)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if got["sub"] != "user_123" {
			t.Fatalf("expected sub=user_123, got %v", got["sub"])
		}
	})

	t.Run("失敗: exp が過去なら token expired", func(t *testing.T) {
		t.Parallel()

		kid := "kid1"
		priv := mustNewRSAKey(t)
		jwks := testJWKS{Keys: []testJWK{jwkFromPublicKey(kid, &priv.PublicKey)}}
		srv := newJWKSServer(t, jwks)
		t.Cleanup(srv.Close)

		v, _ := NewClerkJWTValidator(srv.URL, "", "")

		claims := map[string]any{
			"sub": "user_123",
			"exp": time.Now().Add(-1 * time.Minute).Unix(),
		}
		token := mustSignRS256JWT(t, kid, claims, priv)

		_, err := v.Verify(context.Background(), token)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("失敗: nbf が未来なら token not yet valid", func(t *testing.T) {
		t.Parallel()

		kid := "kid1"
		priv := mustNewRSAKey(t)
		jwks := testJWKS{Keys: []testJWK{jwkFromPublicKey(kid, &priv.PublicKey)}}
		srv := newJWKSServer(t, jwks)
		t.Cleanup(srv.Close)

		v, _ := NewClerkJWTValidator(srv.URL, "", "")

		claims := map[string]any{
			"sub": "user_123",
			"nbf": time.Now().Add(1 * time.Minute).Unix(),
			"exp": time.Now().Add(10 * time.Minute).Unix(),
		}
		token := mustSignRS256JWT(t, kid, claims, priv)

		_, err := v.Verify(context.Background(), token)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("失敗: issuer 不一致", func(t *testing.T) {
		t.Parallel()

		kid := "kid1"
		priv := mustNewRSAKey(t)
		jwks := testJWKS{Keys: []testJWK{jwkFromPublicKey(kid, &priv.PublicKey)}}
		srv := newJWKSServer(t, jwks)
		t.Cleanup(srv.Close)

		v, _ := NewClerkJWTValidator(srv.URL, "iss-ok", "")

		claims := map[string]any{
			"iss": "iss-ng",
			"sub": "user_123",
			"exp": time.Now().Add(10 * time.Minute).Unix(),
		}
		token := mustSignRS256JWT(t, kid, claims, priv)

		_, err := v.Verify(context.Background(), token)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("失敗: audience 不一致", func(t *testing.T) {
		t.Parallel()

		kid := "kid1"
		priv := mustNewRSAKey(t)
		jwks := testJWKS{Keys: []testJWK{jwkFromPublicKey(kid, &priv.PublicKey)}}
		srv := newJWKSServer(t, jwks)
		t.Cleanup(srv.Close)

		v, _ := NewClerkJWTValidator(srv.URL, "", "aud-ok")

		claims := map[string]any{
			"aud": "aud-ng",
			"sub": "user_123",
			"exp": time.Now().Add(10 * time.Minute).Unix(),
		}
		token := mustSignRS256JWT(t, kid, claims, priv)

		_, err := v.Verify(context.Background(), token)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("失敗: alg が RS256 以外", func(t *testing.T) {
		t.Parallel()

		header := map[string]any{"alg": "HS256", "kid": "kid1", "typ": "JWT"}
		claims := map[string]any{"sub": "user_123"}
		token := mustBuildJWTWithHeader(t, header, claims, "")

		v, _ := NewClerkJWTValidator("http://example.invalid/jwks", "", "")
		_, err := v.Verify(context.Background(), token)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("失敗: kid が無い", func(t *testing.T) {
		t.Parallel()

		header := map[string]any{"alg": "RS256", "kid": "", "typ": "JWT"}
		claims := map[string]any{"sub": "user_123"}
		token := mustBuildJWTWithHeader(t, header, claims, "")

		v, _ := NewClerkJWTValidator("http://example.invalid/jwks", "", "")
		_, err := v.Verify(context.Background(), token)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("失敗: 署名が不正", func(t *testing.T) {
		t.Parallel()

		kid := "kid1"
		priv := mustNewRSAKey(t)
		other := mustNewRSAKey(t)

		jwks := testJWKS{Keys: []testJWK{jwkFromPublicKey(kid, &priv.PublicKey)}}
		srv := newJWKSServer(t, jwks)
		t.Cleanup(srv.Close)

		v, _ := NewClerkJWTValidator(srv.URL, "", "")

		claims := map[string]any{
			"sub": "user_123",
			"exp": time.Now().Add(10 * time.Minute).Unix(),
		}
		// JWKSとは別の秘密鍵で署名して不正化
		token := mustSignRS256JWT(t, kid, claims, other)

		_, err := v.Verify(context.Background(), token)
		if err == nil {
			t.Fatalf("expected error")
		}
	})
}

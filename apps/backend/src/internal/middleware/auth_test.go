package middleware

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/service"
	"cinetag-backend/src/internal/testutil"

	"github.com/gin-gonic/gin"
)

type fakeUserService struct {
	EnsureUserFn         func(ctx context.Context, clerkUser service.ClerkUserInfo) (*model.User, error)
	GetUserByDisplayIDFn func(ctx context.Context, displayID string) (*model.User, error)
}

func (f *fakeUserService) EnsureUser(ctx context.Context, clerkUser service.ClerkUserInfo) (*model.User, error) {
	if f.EnsureUserFn == nil {
		return &model.User{ID: "u1"}, nil
	}
	return f.EnsureUserFn(ctx, clerkUser)
}

func (f *fakeUserService) GetUserByDisplayID(ctx context.Context, displayID string) (*model.User, error) {
	if f.GetUserByDisplayIDFn == nil {
		return nil, errors.New("not found")
	}
	return f.GetUserByDisplayIDFn(ctx, displayID)
}

func (f *fakeUserService) FollowUser(ctx context.Context, followerID, followeeID string) error {
	return nil
}

func (f *fakeUserService) UnfollowUser(ctx context.Context, followerID, followeeID string) error {
	return nil
}

func (f *fakeUserService) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	return false, nil
}

func (f *fakeUserService) ListFollowing(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
	return []*model.User{}, 0, nil
}

func (f *fakeUserService) ListFollowers(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
	return []*model.User{}, 0, nil
}

func (f *fakeUserService) GetFollowStats(ctx context.Context, userID string) (following int64, followers int64, err error) {
	return 0, 0, nil
}

func (f *fakeUserService) HandleClerkUserDeleted(ctx context.Context, clerkUserID string) error {
	return nil
}

func newAuthTestRouter(t *testing.T, mw gin.HandlerFunc) *gin.Engine {
	t.Helper()

	r := testutil.NewTestRouter()
	r.Use(mw)
	r.GET("/ok", func(c *gin.Context) {
		userVal, _ := c.Get("user")
		if userVal == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user missing"})
			return
		}
		u, ok := userVal.(*model.User)
		if !ok || u == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user invalid"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": u.ID})
	})
	return r
}

func TestAuthMiddleware(t *testing.T) {
	t.Run("設定不備: CLERK_JWKS_URL が空なら 500", func(t *testing.T) {
		t.Setenv("CLERK_JWKS_URL", "")
		mw := NewAuthMiddleware(&fakeUserService{})
		r := newAuthTestRouter(t, mw)

		rw := testutil.PerformRequest(r, http.MethodGet, "/ok", nil, map[string]string{
			"Authorization": "Bearer abc.def.ghi",
		})
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("未認証: Authorization ヘッダなしは 401", func(t *testing.T) {
		// Verify に到達しないのでURLはダミーでよい
		t.Setenv("CLERK_JWKS_URL", "http://example.invalid/jwks")
		mw := NewAuthMiddleware(&fakeUserService{})
		r := newAuthTestRouter(t, mw)

		rw := testutil.PerformRequest(r, http.MethodGet, "/ok", nil, nil)
		if rw.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("未認証: Bearer 形式でない場合は 401", func(t *testing.T) {
		t.Setenv("CLERK_JWKS_URL", "http://example.invalid/jwks")
		mw := NewAuthMiddleware(&fakeUserService{})
		r := newAuthTestRouter(t, mw)

		rw := testutil.PerformRequest(r, http.MethodGet, "/ok", nil, map[string]string{
			"Authorization": "Basic xxx",
		})
		if rw.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("未認証: トークン形式が不正なら 401", func(t *testing.T) {
		t.Setenv("CLERK_JWKS_URL", "http://example.invalid/jwks")
		mw := NewAuthMiddleware(&fakeUserService{})
		r := newAuthTestRouter(t, mw)

		rw := testutil.PerformRequest(r, http.MethodGet, "/ok", nil, map[string]string{
			"Authorization": "Bearer abc",
		})
		if rw.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("未認証: sub が無いなら 401", func(t *testing.T) {
		kid := "kid1"
		priv := mustNewRSAKey(t)
		jwks := testJWKS{Keys: []testJWK{jwkFromPublicKey(kid, &priv.PublicKey)}}
		srv := newJWKSServer(t, jwks)
		t.Cleanup(srv.Close)

		t.Setenv("CLERK_JWKS_URL", srv.URL)

		claims := map[string]any{
			// "sub" を入れない
			"email": "a@example.com",
			"exp":   time.Now().Add(10 * time.Minute).Unix(),
		}
		token := mustSignRS256JWT(t, kid, claims, priv)

		mw := NewAuthMiddleware(&fakeUserService{})
		r := newAuthTestRouter(t, mw)
		rw := testutil.PerformRequest(r, http.MethodGet, "/ok", nil, map[string]string{
			"Authorization": "Bearer " + token,
		})
		if rw.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("未認証: email が無いなら 401", func(t *testing.T) {
		kid := "kid1"
		priv := mustNewRSAKey(t)
		jwks := testJWKS{Keys: []testJWK{jwkFromPublicKey(kid, &priv.PublicKey)}}
		srv := newJWKSServer(t, jwks)
		t.Cleanup(srv.Close)

		t.Setenv("CLERK_JWKS_URL", srv.URL)

		claims := map[string]any{
			"sub": "user_123",
			// "email" を入れない
			"exp": time.Now().Add(10 * time.Minute).Unix(),
		}
		token := mustSignRS256JWT(t, kid, claims, priv)

		mw := NewAuthMiddleware(&fakeUserService{})
		r := newAuthTestRouter(t, mw)
		rw := testutil.PerformRequest(r, http.MethodGet, "/ok", nil, map[string]string{
			"Authorization": "Bearer " + token,
		})
		if rw.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("ユーザー同期が失敗: 500", func(t *testing.T) {
		kid := "kid1"
		priv := mustNewRSAKey(t)
		jwks := testJWKS{Keys: []testJWK{jwkFromPublicKey(kid, &priv.PublicKey)}}
		srv := newJWKSServer(t, jwks)
		t.Cleanup(srv.Close)

		t.Setenv("CLERK_JWKS_URL", srv.URL)

		claims := map[string]any{
			"sub":   "user_123",
			"email": "a@example.com",
			"exp":   time.Now().Add(10 * time.Minute).Unix(),
		}
		token := mustSignRS256JWT(t, kid, claims, priv)

		expected := errors.New("db error")
		us := &fakeUserService{EnsureUserFn: func(ctx context.Context, clerkUser service.ClerkUserInfo) (*model.User, error) {
			return nil, expected
		}}

		mw := NewAuthMiddleware(us)
		r := newAuthTestRouter(t, mw)
		rw := testutil.PerformRequest(r, http.MethodGet, "/ok", nil, map[string]string{
			"Authorization": "Bearer " + token,
		})
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("成功: user が context にセットされ後続へ進む", func(t *testing.T) {
		kid := "kid1"
		priv := mustNewRSAKey(t)
		jwks := testJWKS{Keys: []testJWK{jwkFromPublicKey(kid, &priv.PublicKey)}}
		srv := newJWKSServer(t, jwks)
		t.Cleanup(srv.Close)

		t.Setenv("CLERK_JWKS_URL", srv.URL)

		claims := map[string]any{
			"sub":        "user_123",
			"email":      "a@example.com",
			"full_name":  "Full Name",
			"first_name": "First",
			"image_url":  "https://example.com/a.png",
			"exp":        time.Now().Add(10 * time.Minute).Unix(),
		}
		token := mustSignRS256JWT(t, kid, claims, priv)

		us := &fakeUserService{EnsureUserFn: func(ctx context.Context, clerkUser service.ClerkUserInfo) (*model.User, error) {
			if clerkUser.ID != "user_123" || clerkUser.Email != "a@example.com" {
				return nil, errors.New("unexpected clerk user")
			}
			return &model.User{ID: "u-local"}, nil
		}}

		mw := NewAuthMiddleware(us)
		r := newAuthTestRouter(t, mw)
		rw := testutil.PerformRequest(r, http.MethodGet, "/ok", nil, map[string]string{
			"Authorization": "Bearer " + token,
		})
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}
	})
}

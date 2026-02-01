package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/service"
	"cinetag-backend/src/internal/testutil"

	"github.com/gin-gonic/gin"
)

type fakeUserService struct {
	GetUserByDisplayIDFn func(ctx context.Context, displayID string) (*model.User, error)
	FollowUserFn         func(ctx context.Context, followerID, followeeID string) error
	UnfollowUserFn       func(ctx context.Context, followerID, followeeID string) error
	IsFollowingFn        func(ctx context.Context, followerID, followeeID string) (bool, error)
	ListFollowingFn      func(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error)
	ListFollowersFn      func(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error)
	GetFollowStatsFn     func(ctx context.Context, userID string) (following int64, followers int64, err error)
}

func (f *fakeUserService) EnsureUser(ctx context.Context, clerkUser service.ClerkUserInfo) (*model.User, error) {
	return nil, nil
}

func (f *fakeUserService) FindUserByClerkUserID(ctx context.Context, clerkUserID string) (*model.User, error) {
	return nil, nil
}

func (f *fakeUserService) GetUserByDisplayID(ctx context.Context, displayID string) (*model.User, error) {
	if f.GetUserByDisplayIDFn == nil {
		return nil, service.ErrUserNotFound
	}
	return f.GetUserByDisplayIDFn(ctx, displayID)
}

func (f *fakeUserService) FollowUser(ctx context.Context, followerID, followeeID string) error {
	if f.FollowUserFn == nil {
		return nil
	}
	return f.FollowUserFn(ctx, followerID, followeeID)
}

func (f *fakeUserService) UnfollowUser(ctx context.Context, followerID, followeeID string) error {
	if f.UnfollowUserFn == nil {
		return nil
	}
	return f.UnfollowUserFn(ctx, followerID, followeeID)
}

func (f *fakeUserService) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	if f.IsFollowingFn == nil {
		return false, nil
	}
	return f.IsFollowingFn(ctx, followerID, followeeID)
}

func (f *fakeUserService) ListFollowing(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
	if f.ListFollowingFn == nil {
		return []*model.User{}, 0, nil
	}
	return f.ListFollowingFn(ctx, userID, page, pageSize)
}

func (f *fakeUserService) ListFollowers(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
	if f.ListFollowersFn == nil {
		return []*model.User{}, 0, nil
	}
	return f.ListFollowersFn(ctx, userID, page, pageSize)
}

func (f *fakeUserService) GetFollowStats(ctx context.Context, userID string) (following int64, followers int64, err error) {
	if f.GetFollowStatsFn == nil {
		return 0, 0, nil
	}
	return f.GetFollowStatsFn(ctx, userID)
}

func (f *fakeUserService) DeactivateUser(ctx context.Context, userID string) error {
	return nil
}

func newUserHandlerRouter(t *testing.T, userSvc service.UserService, tagSvc service.TagService, user *model.User) *gin.Engine {
	t.Helper()

	r := testutil.NewTestRouter()
	logger := testutil.NewTestLogger()
	h := NewUserHandler(logger, userSvc, tagSvc)

	api := r.Group("/api/v1")

	// 認証が必要なエンドポイント
	auth := api.Group("/")
	if user != nil {
		auth.Use(func(c *gin.Context) {
			c.Set("user", user)
			c.Next()
		})
	}
	auth.GET("/users/me", h.GetMe)
	auth.POST("/users/:displayId/follow", h.FollowUser)
	auth.DELETE("/users/:displayId/follow", h.UnfollowUser)

	// Optional Auth (認証なしでもアクセス可能)
	optionalAuth := api.Group("/")
	if user != nil {
		optionalAuth.Use(func(c *gin.Context) {
			c.Set("user", user)
			c.Next()
		})
	}
	optionalAuth.GET("/users/:displayId", h.GetUserByDisplayID)
	optionalAuth.GET("/users/:displayId/tags", h.ListUserTags)
	optionalAuth.GET("/users/:displayId/following", h.ListFollowing)
	optionalAuth.GET("/users/:displayId/followers", h.ListFollowers)
	optionalAuth.GET("/users/:displayId/follow-stats", h.GetUserFollowStats)

	return r
}

func TestUserHandler_GetMe(t *testing.T) {
	t.Parallel()

	t.Run("未認証(user無し): 401", func(t *testing.T) {
		t.Parallel()

		r := newUserHandlerRouter(t, &fakeUserService{}, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/me", nil, nil)
		if rw.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("user が無効な型: 500", func(t *testing.T) {
		t.Parallel()

		r := testutil.NewTestRouter()
		logger := testutil.NewTestLogger()
		h := NewUserHandler(logger, &fakeUserService{}, &fakeTagService{})

		r.Use(func(c *gin.Context) {
			c.Set("user", "invalid-user-type")
			c.Next()
		})
		r.GET("/api/v1/users/me", h.GetMe)

		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/me", nil, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("成功: 200 かつユーザー情報が返る", func(t *testing.T) {
		t.Parallel()

		avatarURL := "https://example.com/avatar.png"
		bio := "Hello, I'm a test user"
		u := &model.User{
			ID:          "u1",
			DisplayID:   "user1",
			DisplayName: "User One",
			AvatarURL:   &avatarURL,
			Bio:         &bio,
		}

		r := newUserHandlerRouter(t, &fakeUserService{}, &fakeTagService{}, u)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/me", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["id"] != "u1" || resp["display_id"] != "user1" || resp["display_name"] != "User One" {
			t.Fatalf("unexpected response: %+v", resp)
		}
		if resp["avatar_url"] != avatarURL {
			t.Fatalf("expected avatar_url=%s, got %v", avatarURL, resp["avatar_url"])
		}
		if resp["bio"] != bio {
			t.Fatalf("expected bio=%s, got %v", bio, resp["bio"])
		}
	})
}

func TestUserHandler_GetUserByDisplayID(t *testing.T) {
	t.Parallel()

	t.Run("display_id が空: 400", func(t *testing.T) {
		t.Parallel()

		r := newUserHandlerRouter(t, &fakeUserService{}, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/", nil, nil)
		if rw.Code != http.StatusNotFound {
			// Ginのルーティングで404になる
			t.Fatalf("expected 404, got %d", rw.Code)
		}
	})

	t.Run("ユーザーが見つからない: 404", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return nil, service.ErrUserNotFound
			},
		}

		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/user1", nil, nil)
		if rw.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rw.Code)
		}
	})

	t.Run("サービスが失敗: 500", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return nil, errors.New("db error")
			},
		}

		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/user1", nil, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("成功: 200 かつユーザー情報が返る", func(t *testing.T) {
		t.Parallel()

		avatarURL := "https://example.com/avatar.png"
		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return &model.User{
					ID:          "u1",
					DisplayID:   displayID,
					DisplayName: "User One",
					AvatarURL:   &avatarURL,
				}, nil
			},
		}

		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/user1", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["display_id"] != "user1" {
			t.Fatalf("unexpected display_id: %v", resp["display_id"])
		}
	})
}

func TestUserHandler_FollowUser(t *testing.T) {
	t.Parallel()

	t.Run("未認証(user無し): 401", func(t *testing.T) {
		t.Parallel()

		r := newUserHandlerRouter(t, &fakeUserService{}, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/users/user2/follow", nil, nil)
		if rw.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("フォロー対象が見つからない: 404", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return nil, service.ErrUserNotFound
			},
		}

		u := &model.User{ID: "u1"}
		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, u)
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/users/user2/follow", nil, nil)
		if rw.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rw.Code)
		}
	})

	t.Run("自分自身をフォロー: 400", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return &model.User{ID: "u1", DisplayID: displayID}, nil
			},
			FollowUserFn: func(ctx context.Context, followerID, followeeID string) error {
				return service.ErrCannotFollowSelf
			},
		}

		u := &model.User{ID: "u1"}
		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, u)
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/users/user1/follow", nil, nil)
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}
	})

	t.Run("既にフォロー済み: 409", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return &model.User{ID: "u2", DisplayID: displayID}, nil
			},
			FollowUserFn: func(ctx context.Context, followerID, followeeID string) error {
				return service.ErrAlreadyFollowing
			},
		}

		u := &model.User{ID: "u1"}
		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, u)
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/users/user2/follow", nil, nil)
		if rw.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rw.Code)
		}
	})

	t.Run("成功: 200 OK", func(t *testing.T) {
		t.Parallel()

		var gotFollowerID, gotFolloweeID string
		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return &model.User{ID: "u2", DisplayID: displayID}, nil
			},
			FollowUserFn: func(ctx context.Context, followerID, followeeID string) error {
				gotFollowerID = followerID
				gotFolloweeID = followeeID
				return nil
			},
		}

		u := &model.User{ID: "u1"}
		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, u)
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/users/user2/follow", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}
		if gotFollowerID != "u1" || gotFolloweeID != "u2" {
			t.Fatalf("unexpected args: followerID=%s followeeID=%s", gotFollowerID, gotFolloweeID)
		}
	})
}

func TestUserHandler_UnfollowUser(t *testing.T) {
	t.Parallel()

	t.Run("未認証(user無し): 401", func(t *testing.T) {
		t.Parallel()

		r := newUserHandlerRouter(t, &fakeUserService{}, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodDelete, "/api/v1/users/user2/follow", nil, nil)
		if rw.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rw.Code)
		}
	})

	t.Run("フォロー対象が見つからない: 404", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return nil, service.ErrUserNotFound
			},
		}

		u := &model.User{ID: "u1"}
		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, u)
		rw := testutil.PerformRequest(r, http.MethodDelete, "/api/v1/users/user2/follow", nil, nil)
		if rw.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rw.Code)
		}
	})

	t.Run("フォローしていない: 409", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return &model.User{ID: "u2", DisplayID: displayID}, nil
			},
			UnfollowUserFn: func(ctx context.Context, followerID, followeeID string) error {
				return service.ErrNotFollowing
			},
		}

		u := &model.User{ID: "u1"}
		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, u)
		rw := testutil.PerformRequest(r, http.MethodDelete, "/api/v1/users/user2/follow", nil, nil)
		if rw.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rw.Code)
		}
	})

	t.Run("成功: 200 OK", func(t *testing.T) {
		t.Parallel()

		var gotFollowerID, gotFolloweeID string
		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return &model.User{ID: "u2", DisplayID: displayID}, nil
			},
			UnfollowUserFn: func(ctx context.Context, followerID, followeeID string) error {
				gotFollowerID = followerID
				gotFolloweeID = followeeID
				return nil
			},
		}

		u := &model.User{ID: "u1"}
		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, u)
		rw := testutil.PerformRequest(r, http.MethodDelete, "/api/v1/users/user2/follow", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}
		if gotFollowerID != "u1" || gotFolloweeID != "u2" {
			t.Fatalf("unexpected args: followerID=%s followeeID=%s", gotFollowerID, gotFolloweeID)
		}
	})
}

func TestUserHandler_ListUserTags(t *testing.T) {
	t.Parallel()

	t.Run("display_id が空: 404", func(t *testing.T) {
		t.Parallel()

		r := newUserHandlerRouter(t, &fakeUserService{}, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/", nil, nil)
		// Ginのルーティングで404になる
		if rw.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rw.Code)
		}
	})

	t.Run("ユーザーが見つからない: 404", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return nil, service.ErrUserNotFound
			},
		}

		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/user1/tags", nil, nil)
		if rw.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rw.Code)
		}
	})

	t.Run("タグサービスが失敗: 500", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return &model.User{ID: "u1", DisplayID: displayID}, nil
			},
		}

		tagSvc := &fakeTagService{
			ListTagsByUserIDFn: func(ctx context.Context, userID string, publicOnly bool, page, pageSize int) ([]service.TagListItem, int64, error) {
				return nil, 0, errors.New("db error")
			},
		}

		r := newUserHandlerRouter(t, userSvc, tagSvc, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/user1/tags", nil, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("成功（publicOnly=true）: 200", func(t *testing.T) {
		t.Parallel()

		var gotUserID string
		var gotPublicOnly bool
		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return &model.User{ID: "u1", DisplayID: displayID}, nil
			},
		}

		tagSvc := &fakeTagService{
			ListTagsByUserIDFn: func(ctx context.Context, userID string, publicOnly bool, page, pageSize int) ([]service.TagListItem, int64, error) {
				gotUserID = userID
				gotPublicOnly = publicOnly
				return []service.TagListItem{
					{ID: "t1", Title: "Tag 1", IsPublic: true, Author: "User One", AuthorDisplayID: "user1"},
				}, 1, nil
			},
		}

		r := newUserHandlerRouter(t, userSvc, tagSvc, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/user1/tags", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}
		if gotUserID != "u1" {
			t.Fatalf("expected userID=u1, got %s", gotUserID)
		}
		if !gotPublicOnly {
			t.Fatalf("expected publicOnly=true, got false")
		}
	})

	t.Run("成功（本人の場合 publicOnly=false）: 200", func(t *testing.T) {
		t.Parallel()

		var gotPublicOnly bool
		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return &model.User{ID: "u1", DisplayID: displayID}, nil
			},
		}

		tagSvc := &fakeTagService{
			ListTagsByUserIDFn: func(ctx context.Context, userID string, publicOnly bool, page, pageSize int) ([]service.TagListItem, int64, error) {
				gotPublicOnly = publicOnly
				return []service.TagListItem{
					{ID: "t1", Title: "Tag 1", IsPublic: true, Author: "User One", AuthorDisplayID: "user1"},
					{ID: "t2", Title: "Tag 2", IsPublic: false, Author: "User One", AuthorDisplayID: "user1"},
				}, 2, nil
			},
		}

		u := &model.User{ID: "u1"}
		r := newUserHandlerRouter(t, userSvc, tagSvc, u)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/user1/tags", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}
		if gotPublicOnly {
			t.Fatalf("expected publicOnly=false, got true")
		}
	})
}

func TestUserHandler_ListFollowing(t *testing.T) {
	t.Parallel()

	t.Run("display_id が空: 400", func(t *testing.T) {
		t.Parallel()

		r := newUserHandlerRouter(t, &fakeUserService{}, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users//following", nil, nil)
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}
	})

	t.Run("ユーザーが見つからない: 404", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return nil, service.ErrUserNotFound
			},
		}

		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/user1/following", nil, nil)
		if rw.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rw.Code)
		}
	})

	t.Run("サービスが失敗: 500", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return &model.User{ID: "u1", DisplayID: displayID}, nil
			},
			ListFollowingFn: func(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
				return nil, 0, errors.New("db error")
			},
		}

		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/user1/following", nil, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("成功: 200", func(t *testing.T) {
		t.Parallel()

		avatarURL := "https://example.com/avatar.png"
		bio := "Test bio"
		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return &model.User{ID: "u1", DisplayID: displayID}, nil
			},
			ListFollowingFn: func(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
				return []*model.User{
					{ID: "u2", DisplayID: "user2", DisplayName: "User 2", AvatarURL: &avatarURL, Bio: &bio},
				}, 1, nil
			},
		}

		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/user1/following", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["total_count"] != float64(1) {
			t.Fatalf("expected total_count=1, got %v", resp["total_count"])
		}
	})
}

func TestUserHandler_ListFollowers(t *testing.T) {
	t.Parallel()

	t.Run("display_id が空: 400", func(t *testing.T) {
		t.Parallel()

		r := newUserHandlerRouter(t, &fakeUserService{}, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users//followers", nil, nil)
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}
	})

	t.Run("ユーザーが見つからない: 404", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return nil, service.ErrUserNotFound
			},
		}

		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/user1/followers", nil, nil)
		if rw.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rw.Code)
		}
	})

	t.Run("サービスが失敗: 500", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return &model.User{ID: "u1", DisplayID: displayID}, nil
			},
			ListFollowersFn: func(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
				return nil, 0, errors.New("db error")
			},
		}

		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/user1/followers", nil, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("成功: 200", func(t *testing.T) {
		t.Parallel()

		avatarURL := "https://example.com/avatar.png"
		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return &model.User{ID: "u1", DisplayID: displayID}, nil
			},
			ListFollowersFn: func(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
				return []*model.User{
					{ID: "u2", DisplayID: "user2", DisplayName: "User 2", AvatarURL: &avatarURL},
				}, 1, nil
			},
		}

		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/user1/followers", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["total_count"] != float64(1) {
			t.Fatalf("expected total_count=1, got %v", resp["total_count"])
		}
	})
}

func TestUserHandler_GetUserFollowStats(t *testing.T) {
	t.Parallel()

	t.Run("display_id が空: 400", func(t *testing.T) {
		t.Parallel()

		r := newUserHandlerRouter(t, &fakeUserService{}, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users//follow-stats", nil, nil)
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}
	})

	t.Run("ユーザーが見つからない: 404", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return nil, service.ErrUserNotFound
			},
		}

		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/user1/follow-stats", nil, nil)
		if rw.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rw.Code)
		}
	})

	t.Run("サービスが失敗: 500", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return &model.User{ID: "u1", DisplayID: displayID}, nil
			},
			GetFollowStatsFn: func(ctx context.Context, userID string) (int64, int64, error) {
				return 0, 0, errors.New("db error")
			},
		}

		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/user1/follow-stats", nil, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("成功（認証なし）: 200", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return &model.User{ID: "u1", DisplayID: displayID}, nil
			},
			GetFollowStatsFn: func(ctx context.Context, userID string) (int64, int64, error) {
				return 10, 20, nil
			},
		}

		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, nil)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/user1/follow-stats", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["following_count"] != float64(10) {
			t.Fatalf("expected following_count=10, got %v", resp["following_count"])
		}
		if resp["followers_count"] != float64(20) {
			t.Fatalf("expected followers_count=20, got %v", resp["followers_count"])
		}
		if resp["is_following"] != false {
			t.Fatalf("expected is_following=false, got %v", resp["is_following"])
		}
	})

	t.Run("成功（認証あり、is_following=true）: 200", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeUserService{
			GetUserByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return &model.User{ID: "u2", DisplayID: displayID}, nil
			},
			GetFollowStatsFn: func(ctx context.Context, userID string) (int64, int64, error) {
				return 10, 20, nil
			},
			IsFollowingFn: func(ctx context.Context, followerID, followeeID string) (bool, error) {
				return true, nil
			},
		}

		u := &model.User{ID: "u1"}
		r := newUserHandlerRouter(t, userSvc, &fakeTagService{}, u)
		rw := testutil.PerformRequest(r, http.MethodGet, "/api/v1/users/user2/follow-stats", nil, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["is_following"] != true {
			t.Fatalf("expected is_following=true, got %v", resp["is_following"])
		}
	})
}

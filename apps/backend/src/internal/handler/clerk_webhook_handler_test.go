package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/service"
	"cinetag-backend/src/internal/testutil"

	"github.com/gin-gonic/gin"
)

type fakeWebhookUserService struct {
	EnsureUserFn            func(ctx context.Context, clerkUser service.ClerkUserInfo) (*model.User, error)
	FindUserByClerkUserIDFn func(ctx context.Context, clerkUserID string) (*model.User, error)
	DeactivateUserFn        func(ctx context.Context, userID string) error
}

func (f *fakeWebhookUserService) EnsureUser(ctx context.Context, clerkUser service.ClerkUserInfo) (*model.User, error) {
	if f.EnsureUserFn == nil {
		return &model.User{ID: "u1"}, nil
	}
	return f.EnsureUserFn(ctx, clerkUser)
}

func (f *fakeWebhookUserService) FindUserByClerkUserID(ctx context.Context, clerkUserID string) (*model.User, error) {
	if f.FindUserByClerkUserIDFn == nil {
		return nil, service.ErrUserNotFound
	}
	return f.FindUserByClerkUserIDFn(ctx, clerkUserID)
}

func (f *fakeWebhookUserService) GetUserByDisplayID(ctx context.Context, displayID string) (*model.User, error) {
	return nil, service.ErrUserNotFound
}

func (f *fakeWebhookUserService) UpdateUser(ctx context.Context, userID string, input service.UpdateUserInput) (*model.User, error) {
	return nil, nil
}

func (f *fakeWebhookUserService) FollowUser(ctx context.Context, followerID, followeeID string) error {
	return nil
}

func (f *fakeWebhookUserService) UnfollowUser(ctx context.Context, followerID, followeeID string) error {
	return nil
}

func (f *fakeWebhookUserService) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	return false, nil
}

func (f *fakeWebhookUserService) ListFollowing(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
	return []*model.User{}, 0, nil
}

func (f *fakeWebhookUserService) ListFollowers(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
	return []*model.User{}, 0, nil
}

func (f *fakeWebhookUserService) GetFollowStats(ctx context.Context, userID string) (int64, int64, error) {
	return 0, 0, nil
}

func (f *fakeWebhookUserService) DeactivateUser(ctx context.Context, userID string) error {
	if f.DeactivateUserFn == nil {
		return nil
	}
	return f.DeactivateUserFn(ctx, userID)
}

func newWebhookHandlerRouter(t *testing.T, userSvc service.UserService) *gin.Engine {
	t.Helper()

	r := testutil.NewTestRouter()
	logger := testutil.NewTestLogger()
	h := NewClerkWebhookHandler(logger, userSvc)

	r.POST("/api/v1/clerk/webhook", h.HandleWebhook)

	return r
}

func TestClerkWebhookHandler_HandleWebhook(t *testing.T) {
	t.Parallel()

	t.Run("無効なJSON: 400", func(t *testing.T) {
		t.Parallel()

		r := newWebhookHandlerRouter(t, &fakeWebhookUserService{})
		body := []byte("{invalid json")
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/clerk/webhook", body, nil)
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["error"] != "invalid webhook payload" {
			t.Fatalf("unexpected error: %v", resp["error"])
		}
	})

	t.Run("未知のイベントタイプ: 200 (無視)", func(t *testing.T) {
		t.Parallel()

		r := newWebhookHandlerRouter(t, &fakeWebhookUserService{})
		body := mustMarshalJSON(t, map[string]any{
			"type": "user.updated",
			"data": map[string]any{},
		})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/clerk/webhook", body, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}
	})
}

func TestClerkWebhookHandler_UserCreated(t *testing.T) {
	t.Parallel()

	t.Run("無効なdata: 400", func(t *testing.T) {
		t.Parallel()

		r := newWebhookHandlerRouter(t, &fakeWebhookUserService{})
		body := mustMarshalJSON(t, map[string]any{
			"type": "user.created",
			"data": "not an object",
		})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/clerk/webhook", body, nil)
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["error"] != "invalid webhook data" {
			t.Fatalf("unexpected error: %v", resp["error"])
		}
	})

	t.Run("ClerkUserInfo構築失敗(idかemailが空): 500", func(t *testing.T) {
		t.Parallel()

		r := newWebhookHandlerRouter(t, &fakeWebhookUserService{})
		body := mustMarshalJSON(t, map[string]any{
			"type": "user.created",
			"data": map[string]any{
				"id":              "",
				"email_addresses": []any{},
			},
		})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/clerk/webhook", body, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}
	})

	t.Run("EnsureUser失敗: 500", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeWebhookUserService{
			EnsureUserFn: func(ctx context.Context, clerkUser service.ClerkUserInfo) (*model.User, error) {
				return nil, errors.New("db error")
			},
		}

		r := newWebhookHandlerRouter(t, userSvc)
		body := mustMarshalJSON(t, map[string]any{
			"type": "user.created",
			"data": map[string]any{
				"id":         "user_123",
				"first_name": "John",
				"last_name":  "Doe",
				"image_url":  "https://example.com/avatar.png",
				"email_addresses": []any{
					map[string]any{"email_address": "john@example.com"},
				},
			},
		})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/clerk/webhook", body, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["error"] != "failed to sync user" {
			t.Fatalf("unexpected error: %v", resp["error"])
		}
	})

	t.Run("成功: 200", func(t *testing.T) {
		t.Parallel()

		var gotClerkUser service.ClerkUserInfo
		userSvc := &fakeWebhookUserService{
			EnsureUserFn: func(ctx context.Context, clerkUser service.ClerkUserInfo) (*model.User, error) {
				gotClerkUser = clerkUser
				return &model.User{ID: "u1"}, nil
			},
		}

		r := newWebhookHandlerRouter(t, userSvc)
		body := mustMarshalJSON(t, map[string]any{
			"type": "user.created",
			"data": map[string]any{
				"id":         "user_123",
				"first_name": "John",
				"last_name":  "Doe",
				"image_url":  "https://example.com/avatar.png",
				"email_addresses": []any{
					map[string]any{"email_address": "john@example.com"},
				},
			},
		})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/clerk/webhook", body, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}

		if gotClerkUser.ID != "user_123" {
			t.Errorf("ID = %q, want %q", gotClerkUser.ID, "user_123")
		}
		if gotClerkUser.Email != "john@example.com" {
			t.Errorf("Email = %q, want %q", gotClerkUser.Email, "john@example.com")
		}
		if gotClerkUser.FirstName != "John" {
			t.Errorf("FirstName = %q, want %q", gotClerkUser.FirstName, "John")
		}
		if gotClerkUser.LastName != "Doe" {
			t.Errorf("LastName = %q, want %q", gotClerkUser.LastName, "Doe")
		}
		if gotClerkUser.AvatarURL == nil || *gotClerkUser.AvatarURL != "https://example.com/avatar.png" {
			t.Errorf("AvatarURL = %v, want %q", gotClerkUser.AvatarURL, "https://example.com/avatar.png")
		}
	})

	t.Run("成功(image_url無し): 200", func(t *testing.T) {
		t.Parallel()

		var gotClerkUser service.ClerkUserInfo
		userSvc := &fakeWebhookUserService{
			EnsureUserFn: func(ctx context.Context, clerkUser service.ClerkUserInfo) (*model.User, error) {
				gotClerkUser = clerkUser
				return &model.User{ID: "u1"}, nil
			},
		}

		r := newWebhookHandlerRouter(t, userSvc)
		body := mustMarshalJSON(t, map[string]any{
			"type": "user.created",
			"data": map[string]any{
				"id":         "user_456",
				"first_name": "Jane",
				"last_name":  "Smith",
				"image_url":  "",
				"email_addresses": []any{
					map[string]any{"email_address": "jane@example.com"},
				},
			},
		})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/clerk/webhook", body, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}

		if gotClerkUser.AvatarURL != nil {
			t.Errorf("AvatarURL = %v, want nil", gotClerkUser.AvatarURL)
		}
	})
}

func TestClerkWebhookHandler_UserDeleted(t *testing.T) {
	t.Parallel()

	t.Run("無効なdata: 400", func(t *testing.T) {
		t.Parallel()

		r := newWebhookHandlerRouter(t, &fakeWebhookUserService{})
		body := mustMarshalJSON(t, map[string]any{
			"type": "user.deleted",
			"data": "not an object",
		})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/clerk/webhook", body, nil)
		if rw.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rw.Code)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["error"] != "invalid webhook data" {
			t.Fatalf("unexpected error: %v", resp["error"])
		}
	})

	t.Run("ユーザーが存在しない(ErrUserNotFound): 200", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeWebhookUserService{
			FindUserByClerkUserIDFn: func(ctx context.Context, clerkUserID string) (*model.User, error) {
				return nil, service.ErrUserNotFound
			},
		}

		r := newWebhookHandlerRouter(t, userSvc)
		body := mustMarshalJSON(t, map[string]any{
			"type": "user.deleted",
			"data": map[string]any{
				"id": "user_nonexistent",
			},
		})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/clerk/webhook", body, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}
	})

	t.Run("FindUserByClerkUserID失敗(DBエラー): 500", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeWebhookUserService{
			FindUserByClerkUserIDFn: func(ctx context.Context, clerkUserID string) (*model.User, error) {
				return nil, errors.New("db error")
			},
		}

		r := newWebhookHandlerRouter(t, userSvc)
		body := mustMarshalJSON(t, map[string]any{
			"type": "user.deleted",
			"data": map[string]any{
				"id": "user_123",
			},
		})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/clerk/webhook", body, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["error"] != "failed to resolve user by clerk user id" {
			t.Fatalf("unexpected error: %v", resp["error"])
		}
	})

	t.Run("DeactivateUser失敗: 500", func(t *testing.T) {
		t.Parallel()

		userSvc := &fakeWebhookUserService{
			FindUserByClerkUserIDFn: func(ctx context.Context, clerkUserID string) (*model.User, error) {
				return &model.User{ID: "u1", ClerkUserID: clerkUserID}, nil
			},
			DeactivateUserFn: func(ctx context.Context, userID string) error {
				return errors.New("deactivation failed")
			},
		}

		r := newWebhookHandlerRouter(t, userSvc)
		body := mustMarshalJSON(t, map[string]any{
			"type": "user.deleted",
			"data": map[string]any{
				"id": "user_123",
			},
		})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/clerk/webhook", body, nil)
		if rw.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rw.Code)
		}

		resp := map[string]any{}
		testutil.MustUnmarshalJSON(t, rw.Body.Bytes(), &resp)
		if resp["error"] != "failed to deactivate user" {
			t.Fatalf("unexpected error: %v", resp["error"])
		}
	})

	t.Run("成功: 200", func(t *testing.T) {
		t.Parallel()

		var gotUserID string
		var gotClerkUserID string
		userSvc := &fakeWebhookUserService{
			FindUserByClerkUserIDFn: func(ctx context.Context, clerkUserID string) (*model.User, error) {
				gotClerkUserID = clerkUserID
				return &model.User{ID: "u1", ClerkUserID: clerkUserID}, nil
			},
			DeactivateUserFn: func(ctx context.Context, userID string) error {
				gotUserID = userID
				return nil
			},
		}

		r := newWebhookHandlerRouter(t, userSvc)
		body := mustMarshalJSON(t, map[string]any{
			"type": "user.deleted",
			"data": map[string]any{
				"id": "user_123",
			},
		})
		rw := testutil.PerformRequest(r, http.MethodPost, "/api/v1/clerk/webhook", body, nil)
		if rw.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rw.Code)
		}

		if gotClerkUserID != "user_123" {
			t.Errorf("clerkUserID = %q, want %q", gotClerkUserID, "user_123")
		}
		if gotUserID != "u1" {
			t.Errorf("userID = %q, want %q", gotUserID, "u1")
		}
	})
}

func mustMarshalJSON(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	return data
}

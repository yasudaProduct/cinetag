package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/repository"

	"gorm.io/gorm"
)

type fakeUserRepo struct {
	FindByIDFn                 func(ctx context.Context, userID string) (*model.User, error)
	FindByClerkUserIDFn        func(ctx context.Context, clerkUserID string) (*model.User, error)
	FindByDisplayIDFn          func(ctx context.Context, displayID string) (*model.User, error)
	CreateFn                   func(ctx context.Context, user *model.User) error
	UpdateForUserDeactivatedFn func(ctx context.Context, userID string, now time.Time, anonymizedEmail string) error
}

func (f *fakeUserRepo) FindByID(ctx context.Context, userID string) (*model.User, error) {
	if f.FindByIDFn == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return f.FindByIDFn(ctx, userID)
}

func (f *fakeUserRepo) FindByClerkUserID(ctx context.Context, clerkUserID string) (*model.User, error) {
	if f.FindByClerkUserIDFn == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return f.FindByClerkUserIDFn(ctx, clerkUserID)
}

func (f *fakeUserRepo) FindByDisplayID(ctx context.Context, displayID string) (*model.User, error) {
	if f.FindByDisplayIDFn == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return f.FindByDisplayIDFn(ctx, displayID)
}

func (f *fakeUserRepo) Create(ctx context.Context, user *model.User) error {
	if f.CreateFn == nil {
		user.ID = "u1"
		return nil
	}
	return f.CreateFn(ctx, user)
}

func (f *fakeUserRepo) UpdateForUserDeactivated(ctx context.Context, userID string, now time.Time, anonymizedEmail string) error {
	if f.UpdateForUserDeactivatedFn == nil {
		return nil
	}
	return f.UpdateForUserDeactivatedFn(ctx, userID, now, anonymizedEmail)
}

type fakeUserFollowerRepo struct {
	CreateFn            func(ctx context.Context, followerID, followeeID string) error
	DeleteFn            func(ctx context.Context, followerID, followeeID string) error
	DeleteAllByUserIDFn func(ctx context.Context, userID string) error
	IsFollowingFn       func(ctx context.Context, followerID, followeeID string) (bool, error)
	ListFollowingFn     func(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error)
	ListFollowersFn     func(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error)
	CountFollowingFn    func(ctx context.Context, userID string) (int64, error)
	CountFollowersFn    func(ctx context.Context, userID string) (int64, error)
}

func (f *fakeUserFollowerRepo) Create(ctx context.Context, followerID, followeeID string) error {
	if f.CreateFn == nil {
		return nil
	}
	return f.CreateFn(ctx, followerID, followeeID)
}

func (f *fakeUserFollowerRepo) Delete(ctx context.Context, followerID, followeeID string) error {
	if f.DeleteFn == nil {
		return nil
	}
	return f.DeleteFn(ctx, followerID, followeeID)
}

func (f *fakeUserFollowerRepo) DeleteAllByUserID(ctx context.Context, userID string) error {
	if f.DeleteAllByUserIDFn == nil {
		return nil
	}
	return f.DeleteAllByUserIDFn(ctx, userID)
}

func (f *fakeUserFollowerRepo) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	if f.IsFollowingFn == nil {
		return false, nil
	}
	return f.IsFollowingFn(ctx, followerID, followeeID)
}

func (f *fakeUserFollowerRepo) ListFollowing(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
	if f.ListFollowingFn == nil {
		return []*model.User{}, 0, nil
	}
	return f.ListFollowingFn(ctx, userID, page, pageSize)
}

func (f *fakeUserFollowerRepo) ListFollowers(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
	if f.ListFollowersFn == nil {
		return []*model.User{}, 0, nil
	}
	return f.ListFollowersFn(ctx, userID, page, pageSize)
}

func (f *fakeUserFollowerRepo) CountFollowing(ctx context.Context, userID string) (int64, error) {
	if f.CountFollowingFn == nil {
		return 0, nil
	}
	return f.CountFollowingFn(ctx, userID)
}

func (f *fakeUserFollowerRepo) CountFollowers(ctx context.Context, userID string) (int64, error) {
	if f.CountFollowersFn == nil {
		return 0, nil
	}
	return f.CountFollowersFn(ctx, userID)
}

func TestUserService_EnsureUser(t *testing.T) {
	t.Parallel()

	t.Run("入力バリデーション: clerk user id が必須", func(t *testing.T) {
		t.Parallel()

		svc := NewUserService(nil, &fakeUserRepo{}, &fakeUserFollowerRepo{}, nil)
		_, err := svc.EnsureUser(context.Background(), ClerkUserInfo{ID: "", Email: "a@example.com"})
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("入力バリデーション: email が必須", func(t *testing.T) {
		t.Parallel()

		svc := NewUserService(nil, &fakeUserRepo{}, &fakeUserFollowerRepo{}, nil)
		_, err := svc.EnsureUser(context.Background(), ClerkUserInfo{ID: "clerk_1", Email: ""})
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("既存ユーザーがいる: Create は呼ばれずそのまま返る", func(t *testing.T) {
		t.Parallel()

		existing := &model.User{ID: "u_exist", ClerkUserID: "clerk_1", Username: "x"}
		var createCalled bool
		repo := &fakeUserRepo{
			FindByClerkUserIDFn: func(ctx context.Context, clerkUserID string) (*model.User, error) {
				return existing, nil
			},
			CreateFn: func(ctx context.Context, user *model.User) error {
				createCalled = true
				return nil
			},
		}
		svc := NewUserService(nil, repo, &fakeUserFollowerRepo{}, nil)

		out, err := svc.EnsureUser(context.Background(), ClerkUserInfo{ID: "clerk_1", Email: "a@example.com"})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if createCalled {
			t.Fatalf("Create should not be called")
		}
		if out == nil || out.ID != "u_exist" {
			t.Fatalf("unexpected output: %+v", out)
		}
	})

	t.Run("検索が失敗: ErrRecordNotFound 以外はそのまま返る", func(t *testing.T) {
		t.Parallel()

		expected := errors.New("db down")
		repo := &fakeUserRepo{
			FindByClerkUserIDFn: func(ctx context.Context, clerkUserID string) (*model.User, error) {
				return nil, expected
			},
		}
		svc := NewUserService(nil, repo, &fakeUserFollowerRepo{}, nil)

		_, err := svc.EnsureUser(context.Background(), ClerkUserInfo{ID: "clerk_1", Email: "a@example.com"})
		if !errors.Is(err, expected) {
			t.Fatalf("expected propagated error, got: %v", err)
		}
	})

	t.Run("新規作成: 見つからない場合は作成して返る（DisplayName優先）", func(t *testing.T) {
		t.Parallel()

		var created *model.User
		repo := &fakeUserRepo{
			FindByClerkUserIDFn: func(ctx context.Context, clerkUserID string) (*model.User, error) {
				return nil, gorm.ErrRecordNotFound
			},
			CreateFn: func(ctx context.Context, user *model.User) error {
				created = &model.User{
					ClerkUserID: user.ClerkUserID,
					Username:    user.Username,
					DisplayName: user.DisplayName,
					Email:       user.Email,
					AvatarURL:   user.AvatarURL,
				}
				user.ID = "u_new"
				return nil
			},
		}
		svc := NewUserService(nil, repo, &fakeUserFollowerRepo{}, nil)

		avatar := "https://example.com/a.png"
		out, err := svc.EnsureUser(context.Background(), ClerkUserInfo{
			ID:        "clerk_1",
			FirstName: "first",
			LastName:  "last",
			Email:     "a@example.com",
			AvatarURL: &avatar,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if out == nil || out.ID != "u_new" {
			t.Fatalf("unexpected output: %+v", out)
		}
		if created == nil {
			t.Fatalf("expected Create to be called")
		}
		if created.ClerkUserID != "clerk_1" || created.Email != "a@example.com" {
			t.Fatalf("unexpected created user: %+v", created)
		}
		// displayName は FirstName + LastName を優先して使う
		if created.Username != "廃止予定" || created.DisplayName != "first last" {
			t.Fatalf("unexpected name fields: %+v", created)
		}
		if created.AvatarURL == nil || *created.AvatarURL != avatar {
			t.Fatalf("expected avatar url to be set")
		}
	})

	t.Run("新規作成: FirstName/LastName が空なら \"名無し\" を使う", func(t *testing.T) {
		t.Parallel()

		var created *model.User
		repo := &fakeUserRepo{
			FindByClerkUserIDFn: func(ctx context.Context, clerkUserID string) (*model.User, error) {
				return nil, gorm.ErrRecordNotFound
			},
			CreateFn: func(ctx context.Context, user *model.User) error {
				created = user
				user.ID = "u_new"
				return nil
			},
		}
		svc := NewUserService(nil, repo, &fakeUserFollowerRepo{}, nil)

		_, err := svc.EnsureUser(context.Background(), ClerkUserInfo{
			ID:        "clerk_1",
			FirstName: "",
			LastName:  "",
			Email:     "a@example.com",
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if created == nil {
			t.Fatalf("expected Create to be called")
		}
		if created.Username != "廃止予定" || created.DisplayName != "名無し" {
			t.Fatalf("expected 名無し, got username=%q displayName=%q", created.Username, created.DisplayName)
		}
	})

	t.Run("新規作成が失敗: Create エラーはそのまま返る", func(t *testing.T) {
		t.Parallel()

		expected := errors.New("insert failed")
		repo := &fakeUserRepo{
			FindByClerkUserIDFn: func(ctx context.Context, clerkUserID string) (*model.User, error) {
				return nil, gorm.ErrRecordNotFound
			},
			CreateFn: func(ctx context.Context, user *model.User) error {
				return expected
			},
		}
		svc := NewUserService(nil, repo, &fakeUserFollowerRepo{}, nil)

		_, err := svc.EnsureUser(context.Background(), ClerkUserInfo{ID: "clerk_1", Email: "a@example.com"})
		if !errors.Is(err, expected) {
			t.Fatalf("expected propagated error, got: %v", err)
		}
	})

	_ = repository.UserRepository(nil)         // compile-time check: fakeUserRepo implements interface
	_ = repository.UserFollowerRepository(nil) // compile-time check: fakeUserFollowerRepo implements interface
}

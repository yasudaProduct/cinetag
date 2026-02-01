package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/repository"
	"cinetag-backend/src/internal/testutil"

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

		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, &fakeUserFollowerRepo{}, nil)
		_, err := svc.EnsureUser(context.Background(), ClerkUserInfo{ID: "", Email: "a@example.com"})
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("入力バリデーション: email が必須", func(t *testing.T) {
		t.Parallel()

		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, &fakeUserFollowerRepo{}, nil)
		_, err := svc.EnsureUser(context.Background(), ClerkUserInfo{ID: "clerk_1", Email: ""})
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("既存ユーザーがいる: Create は呼ばれずそのまま返る", func(t *testing.T) {
		t.Parallel()

		existing := &model.User{ID: "u_exist", ClerkUserID: "clerk_1", DisplayName: "x"}
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
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, repo, &fakeUserFollowerRepo{}, nil)

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
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, repo, &fakeUserFollowerRepo{}, nil)

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
					DisplayName: user.DisplayName,
					Email:       user.Email,
					AvatarURL:   user.AvatarURL,
				}
				user.ID = "u_new"
				return nil
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, repo, &fakeUserFollowerRepo{}, nil)

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
		if created.DisplayName != "first last" {
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
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, repo, &fakeUserFollowerRepo{}, nil)

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
		if created.DisplayName != "名無し" {
			t.Fatalf("expected 名無し, got displayName=%q", created.DisplayName)
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
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, repo, &fakeUserFollowerRepo{}, nil)

		_, err := svc.EnsureUser(context.Background(), ClerkUserInfo{ID: "clerk_1", Email: "a@example.com"})
		if !errors.Is(err, expected) {
			t.Fatalf("expected propagated error, got: %v", err)
		}
	})

	_ = repository.UserRepository(nil)         // compile-time check: fakeUserRepo implements interface
	_ = repository.UserFollowerRepository(nil) // compile-time check: fakeUserFollowerRepo implements interface
}

func TestUserService_FindUserByClerkUserID(t *testing.T) {
	t.Parallel()

	t.Run("入力バリデーション: clerk_user_id が必須", func(t *testing.T) {
		t.Parallel()
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, &fakeUserFollowerRepo{}, nil)
		_, err := svc.FindUserByClerkUserID(context.Background(), "")
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("ユーザーが見つからない: gorm.ErrRecordNotFound は ErrUserNotFound に変換される", func(t *testing.T) {
		t.Parallel()
		repo := &fakeUserRepo{
			FindByClerkUserIDFn: func(ctx context.Context, clerkUserID string) (*model.User, error) {
				return nil, gorm.ErrRecordNotFound
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, repo, &fakeUserFollowerRepo{}, nil)

		_, err := svc.FindUserByClerkUserID(context.Background(), "clerk_1")
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got: %v", err)
		}
	})

	t.Run("検索が失敗: FindByClerkUserID のエラーはそのまま返る", func(t *testing.T) {
		t.Parallel()
		expected := errors.New("db down")
		repo := &fakeUserRepo{
			FindByClerkUserIDFn: func(ctx context.Context, clerkUserID string) (*model.User, error) {
				return nil, expected
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, repo, &fakeUserFollowerRepo{}, nil)

		_, err := svc.FindUserByClerkUserID(context.Background(), "clerk_1")
		if !errors.Is(err, expected) {
			t.Fatalf("expected propagated error, got: %v", err)
		}
	})

	t.Run("成功: ユーザーが見つかる", func(t *testing.T) {
		t.Parallel()
		expected := &model.User{ID: "u1", ClerkUserID: "clerk_1", DisplayName: "User1"}
		repo := &fakeUserRepo{
			FindByClerkUserIDFn: func(ctx context.Context, clerkUserID string) (*model.User, error) {
				return expected, nil
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, repo, &fakeUserFollowerRepo{}, nil)

		user, err := svc.FindUserByClerkUserID(context.Background(), "clerk_1")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if user == nil || user.ID != "u1" {
			t.Fatalf("unexpected user: %+v", user)
		}
	})
}

func TestUserService_GetUserByDisplayID(t *testing.T) {
	t.Parallel()

	t.Run("入力バリデーション: display_id が必須", func(t *testing.T) {
		t.Parallel()
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, &fakeUserFollowerRepo{}, nil)
		_, err := svc.GetUserByDisplayID(context.Background(), "")
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("ユーザーが見つからない: gorm.ErrRecordNotFound は ErrUserNotFound に変換される", func(t *testing.T) {
		t.Parallel()
		repo := &fakeUserRepo{
			FindByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return nil, gorm.ErrRecordNotFound
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, repo, &fakeUserFollowerRepo{}, nil)

		_, err := svc.GetUserByDisplayID(context.Background(), "user1")
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got: %v", err)
		}
	})

	t.Run("削除済みユーザー: ErrUserNotFound を返す", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		repo := &fakeUserRepo{
			FindByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return &model.User{ID: "u1", DisplayID: displayID, DeletedAt: &now}, nil
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, repo, &fakeUserFollowerRepo{}, nil)

		_, err := svc.GetUserByDisplayID(context.Background(), "user1")
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got: %v", err)
		}
	})

	t.Run("成功: ユーザーが見つかる", func(t *testing.T) {
		t.Parallel()
		expected := &model.User{ID: "u1", DisplayID: "user1", DisplayName: "User1"}
		repo := &fakeUserRepo{
			FindByDisplayIDFn: func(ctx context.Context, displayID string) (*model.User, error) {
				return expected, nil
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, repo, &fakeUserFollowerRepo{}, nil)

		user, err := svc.GetUserByDisplayID(context.Background(), "user1")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if user == nil || user.ID != "u1" {
			t.Fatalf("unexpected user: %+v", user)
		}
	})
}

func TestUserService_FollowUser(t *testing.T) {
	t.Parallel()

	t.Run("入力バリデーション: follower_id が必須", func(t *testing.T) {
		t.Parallel()
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, &fakeUserFollowerRepo{}, nil)
		err := svc.FollowUser(context.Background(), "", "u2")
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("入力バリデーション: followee_id が必須", func(t *testing.T) {
		t.Parallel()
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, &fakeUserFollowerRepo{}, nil)
		err := svc.FollowUser(context.Background(), "u1", "")
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("自分自身をフォロー: ErrCannotFollowSelf", func(t *testing.T) {
		t.Parallel()
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, &fakeUserFollowerRepo{}, nil)
		err := svc.FollowUser(context.Background(), "u1", "u1")
		if !errors.Is(err, ErrCannotFollowSelf) {
			t.Fatalf("expected ErrCannotFollowSelf, got: %v", err)
		}
	})

	t.Run("フォロー対象が存在しない: ErrUserNotFound", func(t *testing.T) {
		t.Parallel()
		repo := &fakeUserRepo{
			FindByIDFn: func(ctx context.Context, userID string) (*model.User, error) {
				return nil, gorm.ErrRecordNotFound
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, repo, &fakeUserFollowerRepo{}, nil)

		err := svc.FollowUser(context.Background(), "u1", "u2")
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got: %v", err)
		}
	})

	t.Run("フォロー対象が削除済み: ErrUserNotFound", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		repo := &fakeUserRepo{
			FindByIDFn: func(ctx context.Context, userID string) (*model.User, error) {
				return &model.User{ID: userID, DeletedAt: &now}, nil
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, repo, &fakeUserFollowerRepo{}, nil)

		err := svc.FollowUser(context.Background(), "u1", "u2")
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got: %v", err)
		}
	})

	t.Run("既にフォロー済み: ErrAlreadyFollowing", func(t *testing.T) {
		t.Parallel()
		repo := &fakeUserRepo{
			FindByIDFn: func(ctx context.Context, userID string) (*model.User, error) {
				return &model.User{ID: userID}, nil
			},
		}
		followerRepo := &fakeUserFollowerRepo{
			IsFollowingFn: func(ctx context.Context, followerID, followeeID string) (bool, error) {
				return true, nil
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, repo, followerRepo, nil)

		err := svc.FollowUser(context.Background(), "u1", "u2")
		if !errors.Is(err, ErrAlreadyFollowing) {
			t.Fatalf("expected ErrAlreadyFollowing, got: %v", err)
		}
	})

	t.Run("成功: フォローが作成される", func(t *testing.T) {
		t.Parallel()
		var gotFollowerID, gotFolloweeID string
		repo := &fakeUserRepo{
			FindByIDFn: func(ctx context.Context, userID string) (*model.User, error) {
				return &model.User{ID: userID}, nil
			},
		}
		followerRepo := &fakeUserFollowerRepo{
			IsFollowingFn: func(ctx context.Context, followerID, followeeID string) (bool, error) {
				return false, nil
			},
			CreateFn: func(ctx context.Context, followerID, followeeID string) error {
				gotFollowerID = followerID
				gotFolloweeID = followeeID
				return nil
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, repo, followerRepo, nil)

		err := svc.FollowUser(context.Background(), "u1", "u2")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if gotFollowerID != "u1" || gotFolloweeID != "u2" {
			t.Fatalf("unexpected args: followerID=%s followeeID=%s", gotFollowerID, gotFolloweeID)
		}
	})
}

func TestUserService_UnfollowUser(t *testing.T) {
	t.Parallel()

	t.Run("入力バリデーション: follower_id が必須", func(t *testing.T) {
		t.Parallel()
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, &fakeUserFollowerRepo{}, nil)
		err := svc.UnfollowUser(context.Background(), "", "u2")
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("入力バリデーション: followee_id が必須", func(t *testing.T) {
		t.Parallel()
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, &fakeUserFollowerRepo{}, nil)
		err := svc.UnfollowUser(context.Background(), "u1", "")
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("フォローしていない: ErrNotFollowing", func(t *testing.T) {
		t.Parallel()
		followerRepo := &fakeUserFollowerRepo{
			IsFollowingFn: func(ctx context.Context, followerID, followeeID string) (bool, error) {
				return false, nil
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, followerRepo, nil)

		err := svc.UnfollowUser(context.Background(), "u1", "u2")
		if !errors.Is(err, ErrNotFollowing) {
			t.Fatalf("expected ErrNotFollowing, got: %v", err)
		}
	})

	t.Run("成功: フォローが削除される", func(t *testing.T) {
		t.Parallel()
		var gotFollowerID, gotFolloweeID string
		followerRepo := &fakeUserFollowerRepo{
			IsFollowingFn: func(ctx context.Context, followerID, followeeID string) (bool, error) {
				return true, nil
			},
			DeleteFn: func(ctx context.Context, followerID, followeeID string) error {
				gotFollowerID = followerID
				gotFolloweeID = followeeID
				return nil
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, followerRepo, nil)

		err := svc.UnfollowUser(context.Background(), "u1", "u2")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if gotFollowerID != "u1" || gotFolloweeID != "u2" {
			t.Fatalf("unexpected args: followerID=%s followeeID=%s", gotFollowerID, gotFolloweeID)
		}
	})
}

func TestUserService_IsFollowing(t *testing.T) {
	t.Parallel()

	t.Run("空のIDの場合: false を返す", func(t *testing.T) {
		t.Parallel()
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, &fakeUserFollowerRepo{}, nil)
		result, err := svc.IsFollowing(context.Background(), "", "u2")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result {
			t.Fatalf("expected false")
		}
	})

	t.Run("フォローしている: true を返す", func(t *testing.T) {
		t.Parallel()
		followerRepo := &fakeUserFollowerRepo{
			IsFollowingFn: func(ctx context.Context, followerID, followeeID string) (bool, error) {
				return true, nil
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, followerRepo, nil)

		result, err := svc.IsFollowing(context.Background(), "u1", "u2")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if !result {
			t.Fatalf("expected true")
		}
	})

	t.Run("フォローしていない: false を返す", func(t *testing.T) {
		t.Parallel()
		followerRepo := &fakeUserFollowerRepo{
			IsFollowingFn: func(ctx context.Context, followerID, followeeID string) (bool, error) {
				return false, nil
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, followerRepo, nil)

		result, err := svc.IsFollowing(context.Background(), "u1", "u2")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result {
			t.Fatalf("expected false")
		}
	})
}

func TestUserService_ListFollowing(t *testing.T) {
	t.Parallel()

	t.Run("入力バリデーション: user_id が必須", func(t *testing.T) {
		t.Parallel()
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, &fakeUserFollowerRepo{}, nil)
		_, _, err := svc.ListFollowing(context.Background(), "", 1, 20)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("ページング正規化: page < 1 はデフォルト 1", func(t *testing.T) {
		t.Parallel()
		var gotPage int
		followerRepo := &fakeUserFollowerRepo{
			ListFollowingFn: func(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
				gotPage = page
				return []*model.User{}, 0, nil
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, followerRepo, nil)

		_, _, err := svc.ListFollowing(context.Background(), "u1", 0, 10)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if gotPage != 1 {
			t.Fatalf("expected page=1, got %d", gotPage)
		}
	})

	t.Run("ページング正規化: page_size > 100 は 20", func(t *testing.T) {
		t.Parallel()
		var gotPageSize int
		followerRepo := &fakeUserFollowerRepo{
			ListFollowingFn: func(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
				gotPageSize = pageSize
				return []*model.User{}, 0, nil
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, followerRepo, nil)

		_, _, err := svc.ListFollowing(context.Background(), "u1", 2, 1000)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if gotPageSize != 20 {
			t.Fatalf("expected pageSize=20, got %d", gotPageSize)
		}
	})

	t.Run("成功: フォロー一覧を返す", func(t *testing.T) {
		t.Parallel()
		expected := []*model.User{
			{ID: "u2", DisplayName: "User2"},
		}
		followerRepo := &fakeUserFollowerRepo{
			ListFollowingFn: func(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
				return expected, 1, nil
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, followerRepo, nil)

		users, total, err := svc.ListFollowing(context.Background(), "u1", 1, 20)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if total != 1 || len(users) != 1 {
			t.Fatalf("unexpected result: total=%d len=%d", total, len(users))
		}
		if users[0].ID != "u2" {
			t.Fatalf("unexpected user: %+v", users[0])
		}
	})
}

func TestUserService_ListFollowers(t *testing.T) {
	t.Parallel()

	t.Run("入力バリデーション: user_id が必須", func(t *testing.T) {
		t.Parallel()
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, &fakeUserFollowerRepo{}, nil)
		_, _, err := svc.ListFollowers(context.Background(), "", 1, 20)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("ページング正規化が反映される", func(t *testing.T) {
		t.Parallel()
		var gotPage, gotPageSize int
		followerRepo := &fakeUserFollowerRepo{
			ListFollowersFn: func(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
				gotPage = page
				gotPageSize = pageSize
				return []*model.User{}, 0, nil
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, followerRepo, nil)

		_, _, err := svc.ListFollowers(context.Background(), "u1", 0, 0)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if gotPage != 1 || gotPageSize != 20 {
			t.Fatalf("expected (1,20), got (%d,%d)", gotPage, gotPageSize)
		}
	})

	t.Run("成功: フォロワー一覧を返す", func(t *testing.T) {
		t.Parallel()
		expected := []*model.User{
			{ID: "u2", DisplayName: "User2"},
			{ID: "u3", DisplayName: "User3"},
		}
		followerRepo := &fakeUserFollowerRepo{
			ListFollowersFn: func(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
				return expected, 2, nil
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, followerRepo, nil)

		users, total, err := svc.ListFollowers(context.Background(), "u1", 1, 20)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if total != 2 || len(users) != 2 {
			t.Fatalf("unexpected result: total=%d len=%d", total, len(users))
		}
	})
}

func TestUserService_GetFollowStats(t *testing.T) {
	t.Parallel()

	t.Run("入力バリデーション: user_id が必須", func(t *testing.T) {
		t.Parallel()
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, &fakeUserFollowerRepo{}, nil)
		_, _, err := svc.GetFollowStats(context.Background(), "")
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("成功: フォロー数とフォロワー数を返す", func(t *testing.T) {
		t.Parallel()
		followerRepo := &fakeUserFollowerRepo{
			CountFollowingFn: func(ctx context.Context, userID string) (int64, error) {
				return 10, nil
			},
			CountFollowersFn: func(ctx context.Context, userID string) (int64, error) {
				return 20, nil
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, followerRepo, nil)

		following, followers, err := svc.GetFollowStats(context.Background(), "u1")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if following != 10 || followers != 20 {
			t.Fatalf("unexpected stats: following=%d followers=%d", following, followers)
		}
	})

	t.Run("CountFollowing が失敗: エラーをそのまま返す", func(t *testing.T) {
		t.Parallel()
		expected := errors.New("count failed")
		followerRepo := &fakeUserFollowerRepo{
			CountFollowingFn: func(ctx context.Context, userID string) (int64, error) {
				return 0, expected
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, followerRepo, nil)

		_, _, err := svc.GetFollowStats(context.Background(), "u1")
		if !errors.Is(err, expected) {
			t.Fatalf("expected propagated error, got: %v", err)
		}
	})

	t.Run("CountFollowers が失敗: エラーをそのまま返す", func(t *testing.T) {
		t.Parallel()
		expected := errors.New("count failed")
		followerRepo := &fakeUserFollowerRepo{
			CountFollowingFn: func(ctx context.Context, userID string) (int64, error) {
				return 10, nil
			},
			CountFollowersFn: func(ctx context.Context, userID string) (int64, error) {
				return 0, expected
			},
		}
		logger := testutil.NewTestLogger()
		svc := NewUserService(logger, nil, &fakeUserRepo{}, followerRepo, nil)

		_, _, err := svc.GetFollowStats(context.Background(), "u1")
		if !errors.Is(err, expected) {
			t.Fatalf("expected propagated error, got: %v", err)
		}
	})
}

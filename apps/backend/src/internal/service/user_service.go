package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/repository"

	"gorm.io/gorm"
)

// ClerkUserInfo は、Clerk 側のユーザー情報のうち、
// バックエンドが users テーブル同期に利用する最小限の情報を表します。
type ClerkUserInfo struct {
	ID        string  // Clerk の user ID
	Email     string  // メールアドレス
	FirstName string  // 名（任意）
	LastName  string  // 姓（任意）
	AvatarURL *string // アイコンURL（任意）
}

// ErrUserNotFound はユーザーが見つからなかった場合のエラーです。
var ErrUserNotFound = errors.New("user not found")

// ErrCannotFollowSelf は自分自身をフォローしようとした場合のエラーです。
var ErrCannotFollowSelf = errors.New("cannot follow yourself")

// ErrAlreadyFollowing は既にフォロー済みの場合のエラーです。
var ErrAlreadyFollowing = errors.New("already following")

// ErrNotFollowing はフォローしていないユーザーをアンフォローしようとした場合のエラーです。
var ErrNotFollowing = errors.New("not following")

// UserService は users テーブルに関するユースケースを表します。
type UserService interface {
	// EnsureUser は Clerk のユーザー情報をもとに、
	// users テーブル上に対応するレコードが存在することを保証します。
	// 既に存在すればそれを返し、存在しなければ新規作成して返します。
	EnsureUser(ctx context.Context, clerkUser ClerkUserInfo) (*model.User, error)

	// GetUserByDisplayID は display_id からユーザー情報を取得します。
	GetUserByDisplayID(ctx context.Context, displayID string) (*model.User, error)

	// FollowUser は指定ユーザーをフォローします。
	FollowUser(ctx context.Context, followerID, followeeID string) error

	// UnfollowUser は指定ユーザーをアンフォローします。
	UnfollowUser(ctx context.Context, followerID, followeeID string) error

	// IsFollowing は followerID が followeeID をフォローしているかチェックします。
	IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error)

	// ListFollowing は指定ユーザーがフォローしているユーザー一覧を取得します。
	ListFollowing(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error)

	// ListFollowers は指定ユーザーをフォローしているユーザー一覧を取得します。
	ListFollowers(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error)

	// GetFollowStats はフォロー数とフォロワー数を取得します。
	GetFollowStats(ctx context.Context, userID string) (following int64, followers int64, err error)
}

type userService struct {
	userRepo         repository.UserRepository
	userFollowerRepo repository.UserFollowerRepository
}

// NewUserService は UserService の実装を生成します。
func NewUserService(userRepo repository.UserRepository, userFollowerRepo repository.UserFollowerRepository) UserService {
	return &userService{
		userRepo:         userRepo,
		userFollowerRepo: userFollowerRepo,
	}
}

// EnsureUser は Clerk ユーザーに対応する users レコードの存在を保証します。
func (s *userService) EnsureUser(ctx context.Context, clerkInfo ClerkUserInfo) (*model.User, error) {
	if clerkInfo.ID == "" {
		return nil, errors.New("clerk user id is required")
	}
	if strings.TrimSpace(clerkInfo.Email) == "" {
		return nil, errors.New("email is required")
	}

	existing, err := s.userRepo.FindByClerkUserID(ctx, clerkInfo.ID)

	switch {
	case err == nil:
		// 既に存在する場合はそのまま返す
		return existing, nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		// それ以外のエラーはそのまま返す
		return nil, err
	}

	displayName := resolveDisplayName(clerkInfo)

	// display_id はランダム生成（重複したら内部で再生成）
	displayID := GenerateUserDisplayID(ctx, s.userRepo)

	user := &model.User{
		ClerkUserID: clerkInfo.ID,
		Username:    "廃止予定",
		DisplayID:   displayID,
		DisplayName: displayName,
		Email:       clerkInfo.Email,
		AvatarURL:   clerkInfo.AvatarURL,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func resolveDisplayName(clerkInfo ClerkUserInfo) string {
	first := strings.TrimSpace(clerkInfo.FirstName)
	last := strings.TrimSpace(clerkInfo.LastName)
	switch {
	case first != "" && last != "":
		return first + " " + last
	case first != "":
		return first
	case last != "":
		return last
	}

	return "名無し"
}

// GetUserByDisplayID は display_id からユーザー情報を取得します。
func (s *userService) GetUserByDisplayID(ctx context.Context, displayID string) (*model.User, error) {
	fmt.Println("[user_service] GetUserByDisplayID", displayID)
	if displayID == "" {
		return nil, errors.New("display_id is required")
	}

	user, err := s.userRepo.FindByDisplayID(ctx, displayID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// FollowUser は指定ユーザーをフォローします。
func (s *userService) FollowUser(ctx context.Context, followerID, followeeID string) error {
	fmt.Println("[user_service] FollowUser", followerID, followeeID)
	if followerID == "" || followeeID == "" {
		return errors.New("follower_id and followee_id are required")
	}

	if followerID == followeeID {
		return ErrCannotFollowSelf
	}

	// フォロー対象のユーザーが存在するか確認
	if _, err := s.userRepo.FindByID(ctx, followeeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// 既にフォロー済みかチェック
	isFollowing, err := s.userFollowerRepo.IsFollowing(ctx, followerID, followeeID)
	if err != nil {
		return err
	}
	if isFollowing {
		return ErrAlreadyFollowing
	}

	return s.userFollowerRepo.Create(ctx, followerID, followeeID)
}

// UnfollowUser は指定ユーザーをアンフォローします。
func (s *userService) UnfollowUser(ctx context.Context, followerID, followeeID string) error {
	if followerID == "" || followeeID == "" {
		return errors.New("follower_id and followee_id are required")
	}

	// フォローしているかチェック
	isFollowing, err := s.userFollowerRepo.IsFollowing(ctx, followerID, followeeID)
	if err != nil {
		return err
	}
	if !isFollowing {
		return ErrNotFollowing
	}

	return s.userFollowerRepo.Delete(ctx, followerID, followeeID)
}

// IsFollowing は followerID が followeeID をフォローしているかチェックします。
func (s *userService) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	if followerID == "" || followeeID == "" {
		return false, nil
	}
	return s.userFollowerRepo.IsFollowing(ctx, followerID, followeeID)
}

// ListFollowing は指定ユーザーがフォローしているユーザー一覧を取得します。
func (s *userService) ListFollowing(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
	if userID == "" {
		return nil, 0, errors.New("user_id is required")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.userFollowerRepo.ListFollowing(ctx, userID, page, pageSize)
}

// ListFollowers は指定ユーザーをフォローしているユーザー一覧を取得します。
func (s *userService) ListFollowers(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
	if userID == "" {
		return nil, 0, errors.New("user_id is required")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.userFollowerRepo.ListFollowers(ctx, userID, page, pageSize)
}

// GetFollowStats はフォロー数とフォロワー数を取得します。
func (s *userService) GetFollowStats(ctx context.Context, userID string) (following int64, followers int64, err error) {
	if userID == "" {
		return 0, 0, errors.New("user_id is required")
	}

	following, err = s.userFollowerRepo.CountFollowing(ctx, userID)
	if err != nil {
		return 0, 0, err
	}

	followers, err = s.userFollowerRepo.CountFollowers(ctx, userID)
	if err != nil {
		return 0, 0, err
	}

	return following, followers, nil
}

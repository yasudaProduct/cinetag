package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/repository"

	"gorm.io/gorm"
)

// Clerk 側のユーザー情報のうち、
// - バックエンドが users テーブル同期に利用する最小限の情報を表す構造体。
type ClerkUserInfo struct {
	ID        string  // Clerk の user ID
	Email     string  // メールアドレス
	FirstName string  // 名（任意）
	LastName  string  // 姓（任意）
	AvatarURL *string // アイコンURL（任意）
}

// ユーザーが見つからなかった場合のエラー。
var ErrUserNotFound = errors.New("user not found")

// 自分自身をフォローしようとした場合のエラー。
var ErrCannotFollowSelf = errors.New("cannot follow yourself")

// 既にフォロー済みの場合のエラー。
var ErrAlreadyFollowing = errors.New("already following")

// フォローしていないユーザーをアンフォローしようとした場合のエラー。
var ErrNotFollowing = errors.New("not following")

// users テーブルに関するユースケースを表すインターフェース。
type UserService interface {
	// Clerk ユーザー情報をもとに、
	// users テーブル上に対応するレコードが存在することを保証する。
	// - 既に存在すればそれを返し、存在しなければ新規作成して返す。
	EnsureUser(ctx context.Context, clerkUser ClerkUserInfo) (*model.User, error)

	// clerk_user_id からユーザー情報を取得する。
	// - 削除済みユーザーも取得対象とする（必要に応じて呼び出し側で扱いを決める）。
	FindUserByClerkUserID(ctx context.Context, clerkUserID string) (*model.User, error)

	// display_id からユーザー情報を取得する。
	GetUserByDisplayID(ctx context.Context, displayID string) (*model.User, error)

	// 指定ユーザーをフォローする。
	FollowUser(ctx context.Context, followerID, followeeID string) error

	// 指定ユーザーをアンフォローする。
	UnfollowUser(ctx context.Context, followerID, followeeID string) error

	// followerID が followeeID をフォローしているかチェックする。
	IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error)

	// 指定ユーザーがフォローしているユーザー一覧を取得する。
	ListFollowing(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error)

	// 指定ユーザーをフォローしているユーザー一覧を取得する。
	ListFollowers(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error)

	// フォロー数とフォロワー数を取得する。
	GetFollowStats(ctx context.Context, userID string) (following int64, followers int64, err error)

	// ユーザーを論理削除＋匿名化し、関連データをクリーンアップする。
	DeactivateUser(ctx context.Context, userID string) error
}

type userService struct {
	db               *gorm.DB
	userRepo         repository.UserRepository
	userFollowerRepo repository.UserFollowerRepository
	tagFollowerRepo  repository.TagFollowerRepository
}

// UserService の実装を生成する。
func NewUserService(db *gorm.DB, userRepo repository.UserRepository, userFollowerRepo repository.UserFollowerRepository, tagFollowerRepo repository.TagFollowerRepository) UserService {
	return &userService{
		db:               db,
		userRepo:         userRepo,
		userFollowerRepo: userFollowerRepo,
		tagFollowerRepo:  tagFollowerRepo,
	}
}

// Clerk ユーザーに対応する users レコードの存在を保証する。
func (s *userService) EnsureUser(ctx context.Context, clerkInfo ClerkUserInfo) (*model.User, error) {
	fmt.Println("[user_service] EnsureUser", clerkInfo.ID)
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

// clerk_user_id からユーザー情報を取得する。
func (s *userService) FindUserByClerkUserID(ctx context.Context, clerkUserID string) (*model.User, error) {
	clerkUserID = strings.TrimSpace(clerkUserID)
	if clerkUserID == "" {
		return nil, errors.New("clerk user id is required")
	}

	u, err := s.userRepo.FindByClerkUserID(ctx, clerkUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

// ユーザー名を解決する。
// - FirstName と LastName が両方存在する場合は FirstName + LastName を返す。
// - FirstName が存在する場合は FirstName を返す。
// - LastName が存在する場合は LastName を返す。
// - どちらも存在しない場合は "名無し" を返す。
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

// display_id からユーザー情報を取得する。
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
	if user != nil && user.DeletedAt != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// 指定ユーザーをフォローする。
func (s *userService) FollowUser(ctx context.Context, followerID, followeeID string) error {
	fmt.Println("[user_service] FollowUser", followerID, followeeID)
	if followerID == "" || followeeID == "" {
		return errors.New("follower_id and followee_id are required")
	}

	if followerID == followeeID {
		return ErrCannotFollowSelf
	}

	// フォロー対象のユーザーが存在するか確認
	followee, err := s.userRepo.FindByID(ctx, followeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if followee != nil && followee.DeletedAt != nil {
		return ErrUserNotFound
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

// 指定ユーザーをアンフォローする。
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

// followerID が followeeID をフォローしているかチェックする。
func (s *userService) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	if followerID == "" || followeeID == "" {
		return false, nil
	}
	return s.userFollowerRepo.IsFollowing(ctx, followerID, followeeID)
}

// 指定ユーザーがフォローしているユーザー一覧を取得する。
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

// 指定ユーザーをフォローしているユーザー一覧を取得する。
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

// フォロー数とフォロワー数を取得する。
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

// ユーザーを、DB側で論理削除＋匿名化し、関連データをクリーンアップする。
func (s *userService) DeactivateUser(ctx context.Context, userID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return errors.New("user id is required")
	}
	if s.db == nil {
		return errors.New("db is required")
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		userRepo := repository.NewUserRepository(tx)
		userFollowerRepo := repository.NewUserFollowerRepository(tx)
		tagFollowerRepo := repository.NewTagFollowerRepository(tx)

		u, err := userRepo.FindByID(ctx, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrUserNotFound
			}
			return err
		}

		// ユーザーを論理削除＋匿名化
		anonymizedEmail := fmt.Sprintf("deleted+%s@example.invalid", u.ID)
		if err := userRepo.UpdateForUserDeactivated(ctx, u.ID, now, anonymizedEmail); err != nil {
			return err
		}

		// 当該ユーザーに紐づくフォロー関係をクリーンアップ
		if err := tagFollowerRepo.DeleteAllByUserID(ctx, u.ID); err != nil {
			return err
		}
		if err := userFollowerRepo.DeleteAllByUserID(ctx, u.ID); err != nil {
			return err
		}

		return nil
	})
}

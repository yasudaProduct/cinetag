package service

import (
	"context"
	"errors"
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

// UserService は users テーブルに関するユースケースを表します。
type UserService interface {
	// EnsureUser は Clerk のユーザー情報をもとに、
	// users テーブル上に対応するレコードが存在することを保証します。
	// 既に存在すればそれを返し、存在しなければ新規作成して返します。
	EnsureUser(ctx context.Context, clerkUser ClerkUserInfo) (*model.User, error)

	// GetUserByDisplayID は display_id からユーザー情報を取得します。
	GetUserByDisplayID(ctx context.Context, displayID string) (*model.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

// NewUserService は UserService の実装を生成します。
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

// EnsureUser は Clerk ユーザーに対応する users レコードの存在を保証します。
func (s *userService) EnsureUser(ctx context.Context, clerkInfo ClerkUserInfo) (*model.User, error) {
	if clerkInfo.ID == "" {
		return nil, errors.New("clerk user id is required")
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

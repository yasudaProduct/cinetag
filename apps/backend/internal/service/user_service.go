package service

import (
	"context"
	"errors"

	"cinetag-backend/internal/model"

	"gorm.io/gorm"
)

// ClerkUserInfo は、Clerk 側のユーザー情報のうち、
// バックエンドが users テーブル同期に利用する最小限の情報を表します。
type ClerkUserInfo struct {
	ID          string  // Clerk の user ID
	Username    string  // ユーザー名（なければ生成する）
	Email       string  // メールアドレス
	DisplayName string  // 表示名（なければ Username を使う）
	AvatarURL   *string // アイコンURL（任意）
}

// UserService は users テーブルに関するユースケースを表します。
type UserService interface {
	// EnsureUser は Clerk のユーザー情報をもとに、
	// users テーブル上に対応するレコードが存在することを保証します。
	// 既に存在すればそれを返し、存在しなければ新規作成して返します。
	EnsureUser(ctx context.Context, clerkUser ClerkUserInfo) (*model.User, error)
}

type userService struct {
	db *gorm.DB
}

// NewUserService は UserService の実装を生成します。
func NewUserService(db *gorm.DB) UserService {
	return &userService{db: db}
}

// EnsureUser は Clerk ユーザーに対応する users レコードの存在を保証します。
func (s *userService) EnsureUser(ctx context.Context, clerkInfo ClerkUserInfo) (*model.User, error) {
	if clerkInfo.ID == "" {
		return nil, errors.New("clerk user id is required")
	}

	var user model.User
	err := s.db.WithContext(ctx).
		Where("clerk_user_id = ?", clerkInfo.ID).
		First(&user).
		Error

	switch {
	case err == nil:
		// 既に存在する場合はそのまま返す
		return &user, nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		// それ以外のエラーはそのまま返す
		return nil, err
	}

	// 見つからなかった場合は新規作成
	username := clerkInfo.DisplayName
	if username == "" {
		username = "未設定"
	}
	displayName := clerkInfo.DisplayName
	if displayName == "" {
		displayName = username
	}

	user = model.User{
		ClerkUserID: clerkInfo.ID,
		Username:    username,
		DisplayName: displayName,
		Email:       clerkInfo.Email,
		AvatarURL:   clerkInfo.AvatarURL,
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

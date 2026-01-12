package repository

import (
	"context"
	"fmt"
	"time"

	"cinetag-backend/src/internal/model"

	"gorm.io/gorm"
)

// UserRepository は users テーブルの永続化処理を表します。
type UserRepository interface {
	FindByID(ctx context.Context, userID string) (*model.User, error)
	FindByClerkUserID(ctx context.Context, clerkUserID string) (*model.User, error)
	FindByDisplayID(ctx context.Context, displayID string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	UpdateForUserDeactivated(ctx context.Context, userID string, now time.Time, anonymizedEmail string) error
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository は UserRepository の実装を生成します。
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByClerkUserID(ctx context.Context, clerkUserID string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("clerk_user_id = ?", clerkUserID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByDisplayID(ctx context.Context, displayID string) (*model.User, error) {
	fmt.Println("[user_repository] FindByDisplayID", displayID)
	var user model.User
	if err := r.db.WithContext(ctx).Where("display_id = ?", displayID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) UpdateForUserDeactivated(ctx context.Context, userID string, now time.Time, anonymizedEmail string) error {
	updates := map[string]any{
		"deleted_at":   now,
		"display_name": "退会済みユーザー",
		"avatar_url":   nil,
		"bio":          nil,
		"email":        anonymizedEmail,
		"updated_at":   now,
	}

	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Updates(updates).
		Error
}

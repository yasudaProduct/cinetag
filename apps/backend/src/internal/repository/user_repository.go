package repository

import (
	"context"

	"cinetag-backend/src/internal/model"

	"gorm.io/gorm"
)

// UserRepository は users テーブルの永続化処理を表します。
type UserRepository interface {
	FindByClerkUserID(ctx context.Context, clerkUserID string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository は UserRepository の実装を生成します。
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByClerkUserID(ctx context.Context, clerkUserID string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("clerk_user_id = ?", clerkUserID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}




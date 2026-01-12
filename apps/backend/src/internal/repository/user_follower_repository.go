package repository

import (
	"context"

	"cinetag-backend/src/internal/model"

	"gorm.io/gorm"
)

// UserFollowerRepository は user_followers テーブルの永続化処理を表します。
type UserFollowerRepository interface {
	// Create はフォロー関係を作成します
	Create(ctx context.Context, followerID, followeeID string) error
	// Delete はフォロー関係を削除します
	Delete(ctx context.Context, followerID, followeeID string) error
	// DeleteAllByUserID は指定ユーザーが関与するフォロー関係を全て削除します。
	// （follower / followee の両方を対象とします）
	DeleteAllByUserID(ctx context.Context, userID string) error
	// IsFollowing は followerID が followeeID をフォローしているかチェックします
	IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error)
	// ListFollowing は指定ユーザーがフォローしているユーザー一覧を返します
	ListFollowing(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error)
	// ListFollowers は指定ユーザーをフォローしているユーザー一覧を返します
	ListFollowers(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error)
	// CountFollowing は指定ユーザーがフォローしているユーザー数を返します
	CountFollowing(ctx context.Context, userID string) (int64, error)
	// CountFollowers は指定ユーザーのフォロワー数を返します
	CountFollowers(ctx context.Context, userID string) (int64, error)
}

type userFollowerRepository struct {
	db *gorm.DB
}

// NewUserFollowerRepository は UserFollowerRepository の実装を生成します。
func NewUserFollowerRepository(db *gorm.DB) UserFollowerRepository {
	return &userFollowerRepository{db: db}
}

func (r *userFollowerRepository) Create(ctx context.Context, followerID, followeeID string) error {
	follow := &model.UserFollower{
		FollowerID: followerID,
		FolloweeID: followeeID,
	}
	return r.db.WithContext(ctx).Create(follow).Error
}

func (r *userFollowerRepository) Delete(ctx context.Context, followerID, followeeID string) error {
	return r.db.WithContext(ctx).
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Delete(&model.UserFollower{}).Error
}

func (r *userFollowerRepository) DeleteAllByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Where("follower_id = ? OR followee_id = ?", userID, userID).
		Delete(&model.UserFollower{}).Error
}

func (r *userFollowerRepository) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserFollower{}).
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userFollowerRepository) ListFollowing(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	offset := (page - 1) * pageSize

	// フォロー中のユーザー数をカウント
	if err := r.db.WithContext(ctx).
		Model(&model.UserFollower{}).
		Where("follower_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// フォロー中のユーザー一覧を取得
	err := r.db.WithContext(ctx).
		Table("users").
		Select("users.*").
		Joins("INNER JOIN user_followers ON users.id = user_followers.followee_id").
		Where("user_followers.follower_id = ?", userID).
		Order("user_followers.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userFollowerRepository) ListFollowers(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	offset := (page - 1) * pageSize

	// フォロワー数をカウント
	if err := r.db.WithContext(ctx).
		Model(&model.UserFollower{}).
		Where("followee_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// フォロワー一覧を取得
	err := r.db.WithContext(ctx).
		Table("users").
		Select("users.*").
		Joins("INNER JOIN user_followers ON users.id = user_followers.follower_id").
		Where("user_followers.followee_id = ?", userID).
		Order("user_followers.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userFollowerRepository) CountFollowing(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserFollower{}).
		Where("follower_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *userFollowerRepository) CountFollowers(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserFollower{}).
		Where("followee_id = ?", userID).
		Count(&count).Error
	return count, err
}

package repository

import (
	"context"

	"cinetag-backend/src/internal/model"

	"gorm.io/gorm"
)

// user_followers テーブルの永続化処理を表すインターフェース。
type UserFollowerRepository interface {
	// フォロー関係を作成する。
	Create(ctx context.Context, followerID, followeeID string) error
	// フォロー関係を削除する。
	Delete(ctx context.Context, followerID, followeeID string) error
	// 指定ユーザーが関与するフォロー関係を全て削除する。
	// （follower / followee の両方を対象とします）
	DeleteAllByUserID(ctx context.Context, userID string) error
	// followerID が followeeID をフォローしているかチェックする。
	IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error)
	// 指定ユーザーがフォローしているユーザー一覧を取得する。
	ListFollowing(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error)
	// 指定ユーザーをフォローしているユーザー一覧を取得する。
	ListFollowers(ctx context.Context, userID string, page, pageSize int) ([]*model.User, int64, error)
	// 指定ユーザーがフォローしているユーザー数を取得する。
	CountFollowing(ctx context.Context, userID string) (int64, error)
	// 指定ユーザーのフォロワー数を取得する。
	CountFollowers(ctx context.Context, userID string) (int64, error)
	// 指定ユーザーをフォローしているユーザーIDの一覧を取得する（通知用軽量クエリ）。
	ListFollowerIDs(ctx context.Context, userID string) ([]string, error)
}

type userFollowerRepository struct {
	db *gorm.DB
}

// UserFollowerRepository の実装を生成する。
func NewUserFollowerRepository(db *gorm.DB) UserFollowerRepository {
	return &userFollowerRepository{db: db}
}

// フォロー関係を作成する。
func (r *userFollowerRepository) Create(ctx context.Context, followerID, followeeID string) error {
	follow := &model.UserFollower{
		FollowerID: followerID,
		FolloweeID: followeeID,
	}
	return r.db.WithContext(ctx).Create(follow).Error
}

// フォロー関係を削除する。
func (r *userFollowerRepository) Delete(ctx context.Context, followerID, followeeID string) error {
	return r.db.WithContext(ctx).
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Delete(&model.UserFollower{}).Error
}

// 指定ユーザーが関与するフォロー関係を全て削除する。
func (r *userFollowerRepository) DeleteAllByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Where("follower_id = ? OR followee_id = ?", userID, userID).
		Delete(&model.UserFollower{}).Error
}

// followerID が followeeID をフォローしているかチェックする。
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

// 指定ユーザーがフォローしているユーザー一覧を取得する。
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

// 指定ユーザーをフォローしているユーザー一覧を取得する。
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

// 指定ユーザーがフォローしているユーザー数を取得する。
func (r *userFollowerRepository) CountFollowing(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserFollower{}).
		Where("follower_id = ?", userID).
		Count(&count).Error
	return count, err
}

// 指定ユーザーをフォローしているユーザーIDの一覧を取得する（通知用軽量クエリ）。
func (r *userFollowerRepository) ListFollowerIDs(ctx context.Context, userID string) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).
		Model(&model.UserFollower{}).
		Where("followee_id = ?", userID).
		Pluck("follower_id", &ids).Error
	return ids, err
}

// 指定ユーザーのフォロワー数を取得する。
func (r *userFollowerRepository) CountFollowers(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserFollower{}).
		Where("followee_id = ?", userID).
		Count(&count).Error
	return count, err
}

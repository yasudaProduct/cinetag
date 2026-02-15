package repository

import (
	"context"

	"cinetag-backend/src/internal/model"

	"gorm.io/gorm"
)

// tag_followers テーブルの永続化処理を表すインターフェース。
type TagFollowerRepository interface {
	// Create はタグフォロー関係を作成します
	Create(ctx context.Context, tagID, userID string) error
	// Delete はタグフォロー関係を削除します
	Delete(ctx context.Context, tagID, userID string) error
	// DeleteAllByUserID は指定ユーザーに紐づくタグフォロー関係を全て削除します。
	DeleteAllByUserID(ctx context.Context, userID string) error
	// userID が tagID をフォローしているかチェックする。
	IsFollowing(ctx context.Context, tagID, userID string) (bool, error)
	// タグをフォローしているユーザー一覧を取得する。
	ListFollowers(ctx context.Context, tagID string, page, pageSize int) ([]*model.User, int64, error)
	// タグのフォロワー数を取得する。
	CountFollowers(ctx context.Context, tagID string) (int64, error)
	// ユーザーがフォローしているタグ一覧を取得する。
	ListFollowingTags(ctx context.Context, userID string, page, pageSize int) ([]TagSummary, int64, error)
}

type tagFollowerRepository struct {
	db *gorm.DB
}

// TagFollowerRepository を生成する。
func NewTagFollowerRepository(db *gorm.DB) TagFollowerRepository {
	return &tagFollowerRepository{db: db}
}

// タグフォロー関係を作成する。
func (r *tagFollowerRepository) Create(ctx context.Context, tagID, userID string) error {
	follow := &model.TagFollower{
		TagID:  tagID,
		UserID: userID,
	}
	return r.db.WithContext(ctx).Create(follow).Error
}

// タグフォロー関係を削除する。
func (r *tagFollowerRepository) Delete(ctx context.Context, tagID, userID string) error {
	return r.db.WithContext(ctx).
		Where("tag_id = ? AND user_id = ?", tagID, userID).
		Delete(&model.TagFollower{}).Error
}

// 指定ユーザーに紐づくタグフォロー関係を全て削除する。
func (r *tagFollowerRepository) DeleteAllByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.TagFollower{}).Error
}

// userID が tagID をフォローしているかチェックする。
func (r *tagFollowerRepository) IsFollowing(ctx context.Context, tagID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.TagFollower{}).
		Where("tag_id = ? AND user_id = ?", tagID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// タグをフォローしているユーザー一覧を取得する。
func (r *tagFollowerRepository) ListFollowers(ctx context.Context, tagID string, page, pageSize int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	offset := (page - 1) * pageSize

	// フォロワー数をカウント
	if err := r.db.WithContext(ctx).
		Model(&model.TagFollower{}).
		Where("tag_id = ?", tagID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// フォロワー一覧を取得
	err := r.db.WithContext(ctx).
		Table("users").
		Select("users.*").
		Joins("INNER JOIN tag_followers ON users.id = tag_followers.user_id").
		Where("tag_followers.tag_id = ?", tagID).
		Order("tag_followers.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// タグのフォロワー数を取得する。
func (r *tagFollowerRepository) CountFollowers(ctx context.Context, tagID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.TagFollower{}).
		Where("tag_id = ?", tagID).
		Count(&count).Error
	return count, err
}

// ユーザーがフォローしているタグ一覧を取得する。
func (r *tagFollowerRepository) ListFollowingTags(ctx context.Context, userID string, page, pageSize int) ([]TagSummary, int64, error) {
	var total int64

	offset := (page - 1) * pageSize

	// フォロー中の公開タグ数をカウント
	if err := r.db.WithContext(ctx).
		Table("tag_followers AS tf").
		Joins("INNER JOIN tags AS t ON t.id = tf.tag_id").
		Where("tf.user_id = ? AND t.is_public = ?", userID, true).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []TagSummary{}, 0, nil
	}

	// フォロー中のタグ一覧を取得（公開タグのみ、サブクエリで movie_count / follower_count を計算）
	var rows []TagSummary
	err := r.db.WithContext(ctx).
		Table("tags AS t").
		Select(`t.id, t.title, t.description, t.cover_image_url, t.is_public,
				(SELECT COUNT(*) FROM tag_movies WHERE tag_id = t.id) AS movie_count,
				(SELECT COUNT(*) FROM tag_followers WHERE tag_id = t.id) AS follower_count,
				t.created_at,
				u.display_name AS author, u.display_id AS author_display_id`).
		Joins("INNER JOIN tag_followers AS tf ON t.id = tf.tag_id").
		Joins("JOIN users AS u ON u.id = t.user_id").
		Where("tf.user_id = ? AND t.is_public = ?", userID, true).
		Order("tf.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(&rows).Error

	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

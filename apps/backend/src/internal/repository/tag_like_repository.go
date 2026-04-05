package repository

import (
	"context"

	"cinetag-backend/src/internal/model"

	"gorm.io/gorm"
)

// tag_likes テーブルの永続化処理を表すインターフェース。
type TagLikeRepository interface {
	// Create はタグいいね関係を作成します。
	Create(ctx context.Context, tagID, userID string) error
	// Delete はタグいいね関係を削除します。
	Delete(ctx context.Context, tagID, userID string) error
	// DeleteAllByUserID は指定ユーザーに紐づくタグいいね関係を全て削除します。
	DeleteAllByUserID(ctx context.Context, userID string) error
	// IsLiking は userID が tagID をいいねしているかチェックします。
	IsLiking(ctx context.Context, tagID, userID string) (bool, error)
	// CountLikes はタグのいいね数を取得します。
	CountLikes(ctx context.Context, tagID string) (int64, error)
	// ListLikedTags はユーザーがいいねしたタグ一覧を取得します（いいね日時の新しい順）。
	ListLikedTags(ctx context.Context, userID string, page, pageSize int) ([]TagSummary, int64, error)
}

type tagLikeRepository struct {
	db *gorm.DB
}

// TagLikeRepository を生成する。
func NewTagLikeRepository(db *gorm.DB) TagLikeRepository {
	return &tagLikeRepository{db: db}
}

// タグいいね関係を作成する。
func (r *tagLikeRepository) Create(ctx context.Context, tagID, userID string) error {
	like := &model.TagLike{
		TagID:  tagID,
		UserID: userID,
	}
	return r.db.WithContext(ctx).Create(like).Error
}

// タグいいね関係を削除する。
func (r *tagLikeRepository) Delete(ctx context.Context, tagID, userID string) error {
	return r.db.WithContext(ctx).
		Where("tag_id = ? AND user_id = ?", tagID, userID).
		Delete(&model.TagLike{}).Error
}

// 指定ユーザーに紐づくタグいいね関係を全て削除する。
func (r *tagLikeRepository) DeleteAllByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.TagLike{}).Error
}

// userID が tagID をいいねしているかチェックする。
func (r *tagLikeRepository) IsLiking(ctx context.Context, tagID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.TagLike{}).
		Where("tag_id = ? AND user_id = ?", tagID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// タグのいいね数を取得する。
func (r *tagLikeRepository) CountLikes(ctx context.Context, tagID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.TagLike{}).
		Where("tag_id = ?", tagID).
		Count(&count).Error
	return count, err
}

// ListLikedTags はユーザーがいいねしたタグ一覧を取得する。
func (r *tagLikeRepository) ListLikedTags(ctx context.Context, userID string, page, pageSize int) ([]TagSummary, int64, error) {
	var total int64
	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).
		Table("tag_likes AS tl").
		Joins("INNER JOIN tags AS t ON t.id = tl.tag_id").
		Where("tl.user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []TagSummary{}, 0, nil
	}

	var rows []TagSummary
	err := r.db.WithContext(ctx).
		Table("tags AS t").
		Select(`t.id, t.title, t.description, t.cover_image_url, t.is_public,
				(SELECT COUNT(*) FROM tag_movies WHERE tag_id = t.id) AS movie_count,
				(SELECT COUNT(*) FROM tag_followers WHERE tag_id = t.id) AS follower_count,
				(SELECT COUNT(*) FROM tag_likes WHERE tag_id = t.id) AS like_count,
				t.created_at,
				u.display_name AS author, u.display_id AS author_display_id`).
		Joins("INNER JOIN tag_likes AS tl ON t.id = tl.tag_id").
		Joins("JOIN users AS u ON u.id = t.user_id").
		Where("tl.user_id = ?", userID).
		Order("tl.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

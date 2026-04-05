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

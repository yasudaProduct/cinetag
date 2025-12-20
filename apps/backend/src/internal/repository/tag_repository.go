package repository

import (
	"context"
	"errors"
	"time"

	"cinetag-backend/src/internal/model"

	"gorm.io/gorm"
)

// TagListFilter は公開タグ一覧取得時のフィルタ条件を表します。
type TagListFilter struct {
	Query  string
	Sort   string
	Offset int
	Limit  int
}

// TagSummary は公開タグ一覧取得時に返す1件分の情報（DB由来部分）です。
type TagSummary struct {
	ID            string    `gorm:"column:id"`
	Title         string    `gorm:"column:title"`
	Description   *string   `gorm:"column:description"`
	CoverImageURL *string   `gorm:"column:cover_image_url"`
	IsPublic      bool      `gorm:"column:is_public"`
	MovieCount    int       `gorm:"column:movie_count"`
	FollowerCount int       `gorm:"column:follower_count"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	Author        string    `gorm:"column:author"`
}

// TagDetailRow はタグ詳細取得時に返すDB由来の情報です。
// owner 情報を users と JOIN して取得します。
type TagDetailRow struct {
	ID            string    `gorm:"column:id"`
	Title         string    `gorm:"column:title"`
	Description   *string   `gorm:"column:description"`
	CoverImageURL *string   `gorm:"column:cover_image_url"`
	IsPublic      bool      `gorm:"column:is_public"`
	MovieCount    int       `gorm:"column:movie_count"`
	FollowerCount int       `gorm:"column:follower_count"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`

	OwnerID          string  `gorm:"column:owner_id"`
	OwnerUsername    string  `gorm:"column:owner_username"`
	OwnerDisplayName string  `gorm:"column:owner_display_name"`
	OwnerAvatarURL   *string `gorm:"column:owner_avatar_url"`
}

// TagRepository はタグに関する永続化処理を表します。
type TagRepository interface {
	Create(ctx context.Context, tag *model.Tag) error
	FindByID(ctx context.Context, id string) (*model.Tag, error)
	FindDetailByID(ctx context.Context, id string) (*TagDetailRow, error)
	UpdateByID(ctx context.Context, id string, patch TagUpdatePatch) error
	IncrementMovieCount(ctx context.Context, id string, delta int) error
	ListPublicTags(ctx context.Context, filter TagListFilter) ([]TagSummary, int64, error)
}

type tagRepository struct {
	db *gorm.DB
}

// TagUpdatePatch は tags テーブルの部分更新に利用します。
// nil のフィールドは更新しません。
type TagUpdatePatch struct {
	Title         *string
	Description   **string
	CoverImageURL **string
	IsPublic      *bool
}

// NewTagRepository は TagRepository の実装を生成します。
func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) Create(ctx context.Context, tag *model.Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

func (r *tagRepository) FindByID(ctx context.Context, id string) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &tag, nil
}

func (r *tagRepository) FindDetailByID(ctx context.Context, id string) (*TagDetailRow, error) {
	var row TagDetailRow
	err := r.db.WithContext(ctx).
		Table((model.Tag{}).TableName()+" AS t").
		Select(`t.id, t.title, t.description, t.cover_image_url, t.is_public,
				t.movie_count, t.follower_count, t.created_at, t.updated_at,
				u.id AS owner_id, u.username AS owner_username, u.display_name AS owner_display_name, u.avatar_url AS owner_avatar_url`).
		Joins("JOIN "+(model.User{}).TableName()+" AS u ON u.id = t.user_id").
		Where("t.id = ?", id).
		Scan(&row).
		Error
	if err != nil {
		return nil, err
	}
	if row.ID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

func (r *tagRepository) UpdateByID(ctx context.Context, id string, patch TagUpdatePatch) error {
	updates := map[string]any{}
	if patch.Title != nil {
		updates["title"] = *patch.Title
	}
	if patch.Description != nil {
		updates["description"] = *patch.Description
	}
	if patch.CoverImageURL != nil {
		updates["cover_image_url"] = *patch.CoverImageURL
	}
	if patch.IsPublic != nil {
		updates["is_public"] = *patch.IsPublic
	}

	if len(updates) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Model(&model.Tag{}).
		Where("id = ?", id).
		Updates(updates).
		Error
}

func (r *tagRepository) IncrementMovieCount(ctx context.Context, id string, delta int) error {
	if delta == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&model.Tag{}).
		Where("id = ?", id).
		UpdateColumn("movie_count", gorm.Expr("movie_count + ?", delta)).
		Error
}

func (r *tagRepository) ListPublicTags(ctx context.Context, filter TagListFilter) ([]TagSummary, int64, error) {
	if filter.Limit <= 0 {
		return []TagSummary{}, 0, nil
	}

	qb := r.db.WithContext(ctx).
		Table((model.Tag{}).TableName()+" AS t").
		Select(`t.id, t.title, t.description, t.cover_image_url, t.is_public,
				t.movie_count, t.follower_count, t.created_at,
				u.username AS author`).
		Joins("JOIN "+(model.User{}).TableName()+" AS u ON u.id = t.user_id").
		Where("t.is_public = ?", true)

	if filter.Query != "" {
		qb = qb.Where("t.title ILIKE ?", "%"+filter.Query+"%")
	}

	var total int64
	if err := qb.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []TagSummary{}, 0, nil
	}

	switch filter.Sort {
	case "recent":
		qb = qb.Order("t.created_at DESC")
	case "movie_count":
		qb = qb.Order("t.movie_count DESC")
	default:
		qb = qb.Order("t.follower_count DESC")
	}

	var rows []TagSummary
	if err := qb.Limit(filter.Limit).Offset(filter.Offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

package repository

import (
	"context"
	"errors"
	"time"

	"cinetag-backend/src/internal/model"

	"gorm.io/gorm"
)

// 公開タグ一覧取得時のフィルタ条件を表す。
type TagListFilter struct {
	Query  string
	Sort   string
	Offset int
	Limit  int
}

// 公開タグ一覧取得時に返す1件分の情報（DB由来部分）を表す。
type TagSummary struct {
	ID              string    `gorm:"column:id"`
	Title           string    `gorm:"column:title"`
	Description     *string   `gorm:"column:description"`
	CoverImageURL   *string   `gorm:"column:cover_image_url"`
	IsPublic        bool      `gorm:"column:is_public"`
	MovieCount      int       `gorm:"column:movie_count"`
	FollowerCount   int       `gorm:"column:follower_count"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	Author          string    `gorm:"column:author"`
	AuthorDisplayID string    `gorm:"column:author_display_id"`
}

// タグ詳細取得時に返すDB由来の情報を表す。
// owner 情報を users と JOIN して取得します。
type TagDetailRow struct {
	ID             string    `gorm:"column:id"`
	Title          string    `gorm:"column:title"`
	Description    *string   `gorm:"column:description"`
	CoverImageURL  *string   `gorm:"column:cover_image_url"`
	IsPublic       bool      `gorm:"column:is_public"`
	AddMoviePolicy string    `gorm:"column:add_movie_policy"`
	MovieCount     int       `gorm:"column:movie_count"`
	FollowerCount  int       `gorm:"column:follower_count"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`

	OwnerID          string  `gorm:"column:owner_id"`
	OwnerDisplayID   string  `gorm:"column:owner_display_id"`
	OwnerDisplayName string  `gorm:"column:owner_display_name"`
	OwnerAvatarURL   *string `gorm:"column:owner_avatar_url"`
}

// ユーザーのタグ一覧取得時のフィルタ条件を表す。
type UserTagListFilter struct {
	UserID        string
	IncludePublic bool // trueなら公開タグのみ、falseなら全て（自分のページ用）
	Offset        int
	Limit         int
}

// タグに関する永続化処理を表すインターフェース。
type TagRepository interface {
	Create(ctx context.Context, tag *model.Tag) error
	FindByID(ctx context.Context, id string) (*model.Tag, error)
	FindDetailByID(ctx context.Context, id string) (*TagDetailRow, error)
	UpdateByID(ctx context.Context, id string, patch TagUpdatePatch) error
	ListPublicTags(ctx context.Context, filter TagListFilter) ([]TagSummary, int64, error)
	ListTagsByUserID(ctx context.Context, filter UserTagListFilter) ([]TagSummary, int64, error)
}

type tagRepository struct {
	db *gorm.DB
}

// tags テーブルの部分更新に利用する。
// nil のフィールドは更新しません。
type TagUpdatePatch struct {
	Title          *string
	Description    **string
	CoverImageURL  **string
	IsPublic       *bool
	AddMoviePolicy *string
}

// TagRepository の実装を生成する。
func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

// タグを作成する。
func (r *tagRepository) Create(ctx context.Context, tag *model.Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

// 指定IDのタグを取得する。
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

// 指定IDのタグの詳細を取得する。
func (r *tagRepository) FindDetailByID(ctx context.Context, id string) (*TagDetailRow, error) {
	var row TagDetailRow
	err := r.db.WithContext(ctx).
		Table((model.Tag{}).TableName()+" AS t").
		Select(`t.id, t.title, t.description, t.cover_image_url, t.is_public, t.add_movie_policy,
				(SELECT COUNT(*) FROM tag_movies WHERE tag_id = t.id) AS movie_count,
				(SELECT COUNT(*) FROM tag_followers WHERE tag_id = t.id) AS follower_count,
				t.created_at, t.updated_at,
				u.id AS owner_id, u.display_id AS owner_display_id,
				u.display_name AS owner_display_name, u.avatar_url AS owner_avatar_url`).
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

// 指定IDのタグを更新する。
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
	if patch.AddMoviePolicy != nil {
		updates["add_movie_policy"] = *patch.AddMoviePolicy
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

// 公開タグ一覧を取得する。
func (r *tagRepository) ListPublicTags(ctx context.Context, filter TagListFilter) ([]TagSummary, int64, error) {
	if filter.Limit <= 0 {
		return []TagSummary{}, 0, nil
	}

	baseQuery := r.db.WithContext(ctx).
		Table((model.Tag{}).TableName()+" AS t").
		Joins("JOIN "+(model.User{}).TableName()+" AS u ON u.id = t.user_id").
		Where("t.is_public = ?", true)

	if filter.Query != "" {
		baseQuery = baseQuery.Where("t.title ILIKE ?", "%"+filter.Query+"%")
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []TagSummary{}, 0, nil
	}

	// Count()はSELECTをCOUNT(*)に置き換えるため、Select句を再指定
	qb := baseQuery.Select(`t.id, t.title, t.description, t.cover_image_url, t.is_public,
				(SELECT COUNT(*) FROM tag_movies WHERE tag_id = t.id) AS movie_count,
				(SELECT COUNT(*) FROM tag_followers WHERE tag_id = t.id) AS follower_count,
				t.created_at,
				u.display_name AS author, u.display_id AS author_display_id`)

	switch filter.Sort {
	case "recent":
		qb = qb.Order("t.created_at DESC")
	case "movie_count":
		qb = qb.Order("movie_count DESC")
	default:
		qb = qb.Order("follower_count DESC")
	}

	var rows []TagSummary
	if err := qb.Limit(filter.Limit).Offset(filter.Offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

// 指定ユーザーのタグ一覧を取得する。
func (r *tagRepository) ListTagsByUserID(ctx context.Context, filter UserTagListFilter) ([]TagSummary, int64, error) {
	if filter.Limit <= 0 {
		return []TagSummary{}, 0, nil
	}

	baseQuery := r.db.WithContext(ctx).
		Table((model.Tag{}).TableName()+" AS t").
		Joins("JOIN "+(model.User{}).TableName()+" AS u ON u.id = t.user_id").
		Where("t.user_id = ?", filter.UserID)

	// 公開タグのみにフィルタ（他ユーザーのページ閲覧時）
	if filter.IncludePublic {
		baseQuery = baseQuery.Where("t.is_public = ?", true)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []TagSummary{}, 0, nil
	}

	// Count()はSELECTをCOUNT(*)に置き換えるため、Select句を再指定
	qb := baseQuery.Select(`t.id, t.title, t.description, t.cover_image_url, t.is_public,
				(SELECT COUNT(*) FROM tag_movies WHERE tag_id = t.id) AS movie_count,
				(SELECT COUNT(*) FROM tag_followers WHERE tag_id = t.id) AS follower_count,
				t.created_at,
				u.display_name AS author, u.display_id AS author_display_id`).
		Order("t.created_at DESC")

	var rows []TagSummary
	if err := qb.Limit(filter.Limit).Offset(filter.Offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

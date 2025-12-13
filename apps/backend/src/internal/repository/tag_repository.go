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

// TagRepository はタグに関する永続化処理を表します。
type TagRepository interface {
	Create(ctx context.Context, tag *model.Tag) error
	FindByID(ctx context.Context, id string) (*model.Tag, error)
	IncrementMovieCount(ctx context.Context, id string, delta int) error
	ListPublicTags(ctx context.Context, filter TagListFilter) ([]TagSummary, int64, error)
}

type tagRepository struct {
	db *gorm.DB
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

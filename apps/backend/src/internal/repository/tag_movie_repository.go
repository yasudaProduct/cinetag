package repository

import (
	"context"

	"cinetag-backend/src/internal/model"

	"gorm.io/gorm"
)

// TagMovieRepository はタグに紐づく映画(TagMovie)に関する永続化処理を表します。
type TagMovieRepository interface {
	// ListRecentByTag は指定したタグに紐づく映画を、追加順(新しい順)で最大 limit 件まで取得します。
	ListRecentByTag(ctx context.Context, tagID string, limit int) ([]model.TagMovie, error)
}

type tagMovieRepository struct {
	db *gorm.DB
}

// NewTagMovieRepository は TagMovieRepository の実装を生成します。
func NewTagMovieRepository(db *gorm.DB) TagMovieRepository {
	return &tagMovieRepository{db: db}
}

func (r *tagMovieRepository) ListRecentByTag(ctx context.Context, tagID string, limit int) ([]model.TagMovie, error) {
	if limit <= 0 {
		return []model.TagMovie{}, nil
	}

	var tagMovies []model.TagMovie
	if err := r.db.WithContext(ctx).
		Where("tag_id = ?", tagID).
		Order("created_at DESC").
		Limit(limit).
		Find(&tagMovies).Error; err != nil {
		return nil, err
	}

	return tagMovies, nil
}


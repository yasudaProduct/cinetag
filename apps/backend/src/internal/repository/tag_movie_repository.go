package repository

import (
	"context"
	"errors"

	"cinetag-backend/src/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrTagMovieAlreadyExists = errors.New("tag movie already exists")

// TagMovieRepository はタグに紐づく映画(TagMovie)に関する永続化処理を表します。
type TagMovieRepository interface {
	// ListRecentByTag は指定したタグに紐づく映画を、追加順(新しい順)で最大 limit 件まで取得します。
	ListRecentByTag(ctx context.Context, tagID string, limit int) ([]model.TagMovie, error)
	// Create はタグに映画を追加します。
	// ユニーク制約違反（tag_movies_unique）の場合は ErrTagMovieAlreadyExists を返します。
	Create(ctx context.Context, tagMovie *model.TagMovie) error
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

func (r *tagMovieRepository) Create(ctx context.Context, tagMovie *model.TagMovie) error {
	// ユニーク制約(tag_movies_unique)は (tag_id, tmdb_movie_id)。
	// 追加済みの場合はエラーにせず DoNothing にして RowsAffected で判定する。
	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tag_id"}, {Name: "tmdb_movie_id"}},
		DoNothing: true,
	}).Create(tagMovie)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrTagMovieAlreadyExists
	}
	return nil
}


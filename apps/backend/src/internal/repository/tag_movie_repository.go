package repository

import (
	"context"
	"errors"
	"time"

	"cinetag-backend/src/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrTagMovieAlreadyExists = errors.New("tag movie already exists")

// タグに紐づく映画(TagMovie)に関する永続化処理を表すインターフェース。
type TagMovieRepository interface {
	// 指定したタグに紐づく映画を、追加順(新しい順)で最大 limit 件まで取得する。
	ListRecentByTag(ctx context.Context, tagID string, limit int) ([]model.TagMovie, error)
	// 指定したタグに紐づく映画を取得する（ページング対応）。
	// movie_cache を LEFT JOIN し、可能なら映画情報も一緒に返す。
	ListByTag(ctx context.Context, tagID string, offset, limit int) ([]TagMovieWithCache, int64, error)
	// タグに映画を追加する。
	// ユニーク制約違反（tag_movies_unique）の場合は ErrTagMovieAlreadyExists を返す。
	Create(ctx context.Context, tagMovie *model.TagMovie) error
	// 指定したIDのタグ映画を取得する。
	FindByID(ctx context.Context, tagMovieID string) (*model.TagMovie, error)
	// 指定したIDのタグ映画を削除する。
	Delete(ctx context.Context, tagMovieID string) error
	// 指定したタグに映画を追加したユーザー（参加者）を取得する。
	// タグ作成者(ownerID)は除外される。
	ListContributorsByTag(ctx context.Context, tagID string, ownerID string, limit int) ([]TagContributor, int64, error)
}

// タグに映画を追加したユーザー情報です。
type TagContributor struct {
	UserID      string  `gorm:"column:user_id"`
	DisplayID   string  `gorm:"column:display_id"`
	DisplayName string  `gorm:"column:display_name"`
	AvatarURL   *string `gorm:"column:avatar_url"`
}

// tag_movies と movie_cache の結合結果を表す。
// cache 側は存在しない可能性があるため nullable を許容する。
type TagMovieWithCache struct {
	ID          string    `gorm:"column:id"`
	TagID       string    `gorm:"column:tag_id"`
	TmdbMovieID int       `gorm:"column:tmdb_movie_id"`
	AddedByUser string    `gorm:"column:added_by_user_id"`
	Note        *string   `gorm:"column:note"`
	Position    int       `gorm:"column:position"`
	CreatedAt   time.Time `gorm:"column:created_at"`

	MovieTitle         *string    `gorm:"column:movie_title"`
	MovieOriginalTitle *string    `gorm:"column:movie_original_title"`
	MoviePosterPath    *string    `gorm:"column:movie_poster_path"`
	MovieReleaseDate   *time.Time `gorm:"column:movie_release_date"`
	MovieVoteAverage   *float64   `gorm:"column:movie_vote_average"`
}

// タグ映画に関する永続化処理を表すインターフェース。
type tagMovieRepository struct {
	db *gorm.DB
}

// TagMovieRepository の実装を生成する。
func NewTagMovieRepository(db *gorm.DB) TagMovieRepository {
	return &tagMovieRepository{db: db}
}

// 指定したタグに紐づく映画を、追加順(新しい順)で最大 limit 件まで取得する。
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

// 指定したタグに紐づく映画を取得する（ページング対応）。
// movie_cache を LEFT JOIN し、可能なら映画情報も一緒に返す。
func (r *tagMovieRepository) ListByTag(ctx context.Context, tagID string, offset, limit int) ([]TagMovieWithCache, int64, error) {
	if limit <= 0 {
		return []TagMovieWithCache{}, 0, nil
	}
	if offset < 0 {
		offset = 0
	}

	// total count（tag_movies の件数）
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&model.TagMovie{}).
		Where("tag_id = ?", tagID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []TagMovieWithCache{}, 0, nil
	}

	// list with cache join
	var rows []TagMovieWithCache
	err := r.db.WithContext(ctx).
		Table((model.TagMovie{}).TableName()+" AS tm").
		Select(`tm.id, tm.tag_id, tm.tmdb_movie_id, tm.added_by_user_id, tm.note, tm.position, tm.created_at,
		        mc.title AS movie_title, mc.original_title AS movie_original_title, mc.poster_path AS movie_poster_path,
		        mc.release_date AS movie_release_date, mc.vote_average AS movie_vote_average`).
		Joins("LEFT JOIN "+(model.MovieCache{}).TableName()+" AS mc ON mc.tmdb_movie_id = tm.tmdb_movie_id").
		Where("tm.tag_id = ?", tagID).
		Order("tm.position ASC, tm.created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// タグに映画を追加する。
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

// 指定したIDのタグ映画を取得する。
func (r *tagMovieRepository) FindByID(ctx context.Context, tagMovieID string) (*model.TagMovie, error) {
	var tagMovie model.TagMovie
	if err := r.db.WithContext(ctx).Where("id = ?", tagMovieID).First(&tagMovie).Error; err != nil {
		return nil, err
	}
	return &tagMovie, nil
}

// 指定したIDのタグ映画を削除する。
func (r *tagMovieRepository) Delete(ctx context.Context, tagMovieID string) error {
	return r.db.WithContext(ctx).Where("id = ?", tagMovieID).Delete(&model.TagMovie{}).Error
}

// 指定したタグに映画を追加したユーザー（参加者）を取得する。
// タグ作成者(ownerID)は除外される。
func (r *tagMovieRepository) ListContributorsByTag(ctx context.Context, tagID string, ownerID string, limit int) ([]TagContributor, int64, error) {
	if limit <= 0 {
		limit = 10
	}

	// サブクエリで distinct なユーザーIDを取得（タグ作成者は除外）
	// total count
	var total int64
	countQuery := r.db.WithContext(ctx).
		Table((model.TagMovie{}).TableName()).
		Select("COUNT(DISTINCT added_by_user_id)").
		Where("tag_id = ?", tagID).
		Where("added_by_user_id != ?", ownerID)
	if err := countQuery.Scan(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []TagContributor{}, 0, nil
	}

	// ユーザー情報を取得（GROUP BYでユニーク化し、最初に追加した日時順でソート）
	var rows []TagContributor
	err := r.db.WithContext(ctx).
		Table((model.TagMovie{}).TableName()+" AS tm").
		Select("u.id AS user_id, u.display_id, u.display_name, u.avatar_url").
		Joins("JOIN "+(model.User{}).TableName()+" AS u ON u.id = tm.added_by_user_id").
		Where("tm.tag_id = ?", tagID).
		Where("tm.added_by_user_id != ?", ownerID).
		Group("u.id, u.display_id, u.display_name, u.avatar_url").
		Order("MIN(tm.created_at) ASC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

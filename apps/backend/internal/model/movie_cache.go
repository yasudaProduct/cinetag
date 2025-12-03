package model

import (
	"time"

	"gorm.io/datatypes"
)

// MovieCache は TMDb から取得した映画情報のキャッシュを表します。
// docs/database-schema.md の movie_cache テーブル定義に対応します。
type MovieCache struct {
	TmdbMovieID   int            `gorm:"type:integer;primaryKey;column:tmdb_movie_id" json:"tmdb_movie_id"`
	Title         string         `gorm:"type:text;not null" json:"title"`
	OriginalTitle *string        `gorm:"type:text;column:original_title" json:"original_title,omitempty"`
	PosterPath    *string        `gorm:"type:text;column:poster_path" json:"poster_path,omitempty"`
	BackdropPath  *string        `gorm:"type:text;column:backdrop_path" json:"backdrop_path,omitempty"`
	ReleaseDate   *time.Time     `gorm:"type:date;column:release_date" json:"release_date,omitempty"`
	VoteAverage   *float64       `gorm:"type:numeric(3,1);column:vote_average" json:"vote_average,omitempty"`
	Overview      *string        `gorm:"type:text" json:"overview,omitempty"`
	Genres        datatypes.JSON `gorm:"type:jsonb" json:"genres,omitempty"`
	Runtime       *int           `gorm:"type:integer" json:"runtime,omitempty"`
	CachedAt      time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP;column:cached_at" json:"cached_at"`
	ExpiresAt     time.Time      `gorm:"type:timestamptz;not null;default:(CURRENT_TIMESTAMP + interval '7 days');column:expires_at" json:"expires_at"`
}

// TableName は対応するテーブル名を返します。
func (MovieCache) TableName() string {
	return "movie_cache"
}



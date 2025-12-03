package model

import "time"

// CategoryMovie はカテゴリに属する映画を表します。
// docs/database-schema.md の category_movies テーブル定義に対応します。
type CategoryMovie struct {
	ID            string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CategoryID    string    `gorm:"type:uuid;not null;column:category_id;uniqueIndex:category_movies_unique" json:"category_id"`
	TmdbMovieID   int       `gorm:"type:integer;not null;column:tmdb_movie_id;uniqueIndex:category_movies_unique" json:"tmdb_movie_id"`
	AddedByUserID string    `gorm:"type:uuid;not null;column:added_by_user_id" json:"added_by_user_id"`
	Note          *string   `gorm:"type:text" json:"note,omitempty"`
	Position      int       `gorm:"type:integer;not null;default:0" json:"position"`
	CreatedAt     time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
}

// TableName は対応するテーブル名を返します。
func (CategoryMovie) TableName() string {
	return "category_movies"
}



package model

import "time"

// TagMovie はタグに属する映画を表します。
// 旧 category_movies テーブルに相当し、tags と映画の関連を表現します。
type TagMovie struct {
	ID          string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	TagID       string    `gorm:"type:uuid;not null;column:tag_id;uniqueIndex:tag_movies_unique" json:"tag_id"`
	TmdbMovieID int       `gorm:"type:integer;not null;column:tmdb_movie_id;uniqueIndex:tag_movies_unique" json:"tmdb_movie_id"`
	AddedByUser string    `gorm:"type:uuid;not null;column:added_by_user_id" json:"added_by_user_id"`
	Note        *string   `gorm:"type:text" json:"note,omitempty"`
	Position    int       `gorm:"type:integer;not null;default:0" json:"position"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
}

// TableName は対応するテーブル名を返します。
func (TagMovie) TableName() string {
	return "tag_movies"
}



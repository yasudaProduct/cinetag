package model

import "time"

// Tag はユーザーが作成する映画タグを表します。
type Tag struct {
	ID            string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID        string    `gorm:"type:uuid;not null;column:user_id" json:"user_id"`
	Title         string    `gorm:"type:text;not null" json:"title"`
	Description   *string   `gorm:"type:text" json:"description,omitempty"`
	CoverImageURL *string   `gorm:"type:text;column:cover_image_url" json:"cover_image_url,omitempty"`
	IsPublic       bool      `gorm:"type:boolean;not null;default:false;column:is_public" json:"is_public"`
	AddMoviePolicy string    `gorm:"type:text;not null;default:'everyone';column:add_movie_policy" json:"add_movie_policy"`
	MovieCount     int       `gorm:"type:integer;not null;default:0;column:movie_count" json:"movie_count"`
	FollowerCount int       `gorm:"type:integer;not null;default:0;column:follower_count" json:"follower_count"`
	CreatedAt     time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP;column:updated_at" json:"updated_at"`
}

// TableName は対応するテーブル名を返します。
func (Tag) TableName() string {
	return "tags"
}

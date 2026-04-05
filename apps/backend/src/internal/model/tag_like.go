package model

import "time"

// TagLike はタグへのいいね関係を表します。
type TagLike struct {
	TagID     string    `gorm:"type:uuid;primaryKey;column:tag_id" json:"tag_id"`
	UserID    string    `gorm:"type:uuid;primaryKey;column:user_id" json:"user_id"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
}

// TableName は対応するテーブル名を返します。
func (TagLike) TableName() string {
	return "tag_likes"
}

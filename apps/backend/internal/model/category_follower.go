package model

import "time"

// CategoryFollower はカテゴリのフォロー関係を表します。
// docs/database-schema.md の category_followers テーブル定義に対応します。
type CategoryFollower struct {
	CategoryID string    `gorm:"type:uuid;primaryKey;column:category_id" json:"category_id"`
	UserID     string    `gorm:"type:uuid;primaryKey;column:user_id" json:"user_id"`
	CreatedAt  time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
}

// TableName は対応するテーブル名を返します。
func (CategoryFollower) TableName() string {
	return "category_followers"
}



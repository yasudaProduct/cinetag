package model

import "time"

// UserFollower はユーザーのフォロー関係を表します。
type UserFollower struct {
	FollowerID string    `gorm:"type:uuid;primaryKey;column:follower_id" json:"follower_id"`
	FolloweeID string    `gorm:"type:uuid;primaryKey;column:followee_id" json:"followee_id"`
	CreatedAt  time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
}

// TableName は対応するテーブル名を返します。
func (UserFollower) TableName() string {
	return "user_followers"
}

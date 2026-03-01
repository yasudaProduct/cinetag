package model

import "time"

// 通知タイプ定数
const (
	NotificationTypeTagMovieAdded           = "tag_movie_added"
	NotificationTypeTagFollowed             = "tag_followed"
	NotificationTypeUserFollowed            = "user_followed"
	NotificationTypeFollowingUserCreatedTag = "following_user_created_tag"
)

// Notification はアプリ内通知を表すドメインモデルです。
type Notification struct {
	ID               string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RecipientUserID  string     `gorm:"type:uuid;not null;column:recipient_user_id" json:"recipient_user_id"`
	ActorUserID      *string    `gorm:"type:uuid;column:actor_user_id" json:"actor_user_id"`
	NotificationType string     `gorm:"type:text;not null;column:notification_type" json:"notification_type"`
	TagID            *string    `gorm:"type:uuid;column:tag_id" json:"tag_id"`
	TagMovieID       *string    `gorm:"type:uuid;column:tag_movie_id" json:"tag_movie_id"`
	IsRead           bool       `gorm:"not null;default:false;column:is_read" json:"is_read"`
	ReadAt           *time.Time `gorm:"type:timestamptz;column:read_at" json:"read_at"`
	CreatedAt        time.Time  `gorm:"type:timestamptz;not null;default:now();column:created_at" json:"created_at"`
}

// TableName は対応するテーブル名を返します。
func (Notification) TableName() string {
	return "notifications"
}

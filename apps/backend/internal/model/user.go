package model

import "time"

// User はサービスのユーザーを表すドメインモデルです。
// docs/database-schema.md の users テーブル定義に対応します。
type User struct {
	ID          string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ClerkUserID string    `gorm:"type:text;not null;uniqueIndex:users_clerk_user_id_key;column:clerk_user_id" json:"clerk_user_id"`
	Username    string    `gorm:"type:text;not null;uniqueIndex:users_username_key" json:"username"`
	DisplayName string    `gorm:"type:text;not null;column:display_name" json:"display_name"`
	Email       string    `gorm:"type:text;not null" json:"email"`
	AvatarURL   *string   `gorm:"type:text;column:avatar_url" json:"avatar_url,omitempty"`
	Bio         *string   `gorm:"type:text" json:"bio,omitempty"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP;column:updated_at" json:"updated_at"`
}

// TableName は対応するテーブル名を返します。
func (User) TableName() string {
	return "users"
}

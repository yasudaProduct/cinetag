package repository

import (
	"context"
	"time"

	"cinetag-backend/src/internal/model"

	"gorm.io/gorm"
)

// NotificationRow は通知一覧取得時の JOIN 結果を格納するフラット構造体。
// notifications LEFT JOIN users(actor) LEFT JOIN tags LEFT JOIN tag_movies LEFT JOIN movie_cache
type NotificationRow struct {
	// notifications
	ID               string    `gorm:"column:id"`
	RecipientUserID  string    `gorm:"column:recipient_user_id"`
	NotificationType string    `gorm:"column:notification_type"`
	IsRead           bool      `gorm:"column:is_read"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	// actor (users)
	ActorUserID      *string `gorm:"column:actor_user_id"`
	ActorDisplayID   *string `gorm:"column:actor_display_id"`
	ActorDisplayName *string `gorm:"column:actor_display_name"`
	ActorAvatarURL   *string `gorm:"column:actor_avatar_url"`
	// tag
	TagID    *string `gorm:"column:tag_id"`
	TagTitle *string `gorm:"column:tag_title"`
	// movie_cache (via tag_movies)
	MovieTitle *string `gorm:"column:movie_title"`
}

// notifications テーブルの永続化処理を表すインターフェース。
type NotificationRepository interface {
	// Create は通知を1件作成する。
	Create(ctx context.Context, notification *model.Notification) error
	// CreateBatch は通知を一括作成する（フォロワー全員への通知等）。
	CreateBatch(ctx context.Context, notifications []*model.Notification) error
	// ListByRecipient は指定ユーザーの通知一覧を新しい順で返す。
	ListByRecipient(ctx context.Context, userID string, page, pageSize int) ([]*NotificationRow, int64, error)
	// CountUnread は未読通知数を返す。
	CountUnread(ctx context.Context, userID string) (int64, error)
	// MarkAsRead は指定の通知を既読にする。recipient_user_id で所有権チェック。
	MarkAsRead(ctx context.Context, notificationID, userID string) error
	// MarkAllAsRead は指定ユーザーの全通知を既読にする。
	MarkAllAsRead(ctx context.Context, userID string) error
}

type notificationRepository struct {
	db *gorm.DB
}

// NotificationRepository を生成する。
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// 通知を1件作成する。
func (r *notificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// 通知を一括作成する。
func (r *notificationRepository) CreateBatch(ctx context.Context, notifications []*model.Notification) error {
	if len(notifications) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(notifications, 100).Error
}

// 指定ユーザーの通知一覧を新しい順で返す。
func (r *notificationRepository) ListByRecipient(ctx context.Context, userID string, page, pageSize int) ([]*NotificationRow, int64, error) {
	var total int64
	offset := (page - 1) * pageSize

	// 総件数
	if err := r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("recipient_user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*NotificationRow{}, 0, nil
	}

	// JOINクエリ
	var rows []*NotificationRow
	err := r.db.WithContext(ctx).
		Table("notifications AS n").
		Select(`n.id, n.recipient_user_id, n.notification_type, n.is_read, n.created_at,
				n.actor_user_id,
				actor.display_id AS actor_display_id,
				actor.display_name AS actor_display_name,
				actor.avatar_url AS actor_avatar_url,
				n.tag_id,
				t.title AS tag_title,
				mc.title AS movie_title`).
		Joins("LEFT JOIN users AS actor ON actor.id = n.actor_user_id").
		Joins("LEFT JOIN tags AS t ON t.id = n.tag_id").
		Joins("LEFT JOIN tag_movies AS tm ON tm.id = n.tag_movie_id").
		Joins("LEFT JOIN movie_cache AS mc ON mc.tmdb_movie_id = tm.tmdb_movie_id").
		Where("n.recipient_user_id = ?", userID).
		Order("n.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(&rows).Error

	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

// 未読通知数を返す（部分インデックスが効く軽量クエリ）。
func (r *notificationRepository) CountUnread(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("recipient_user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// 指定の通知を既読にする。recipient_user_id で所有権チェック。
func (r *notificationRepository) MarkAsRead(ctx context.Context, notificationID, userID string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("id = ? AND recipient_user_id = ?", notificationID, userID).
		Updates(map[string]any{
			"is_read": true,
			"read_at": now,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// 指定ユーザーの全未読通知を既読にする。
func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("recipient_user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]any{
			"is_read": true,
			"read_at": now,
		}).Error
}

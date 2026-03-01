package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/repository"

	"gorm.io/gorm"
)

// 通知が見つからなかった場合のエラー。
var ErrNotificationNotFound = errors.New("notification not found")

// 通知一覧APIのレスポンスDTO。
type NotificationItem struct {
	ID               string                     `json:"id"`
	NotificationType string                     `json:"notification_type"`
	IsRead           bool                       `json:"is_read"`
	CreatedAt        time.Time                  `json:"created_at"`
	Actor            *ActorSummary              `json:"actor"`
	Tag              *TagSummaryForNotification `json:"tag,omitempty"`
	MovieTitle       *string                    `json:"movie_title,omitempty"`
}

// 通知内のアクター情報。
type ActorSummary struct {
	ID          string  `json:"id"`
	DisplayID   string  `json:"display_id"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

// 通知内のタグ情報。
type TagSummaryForNotification struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// 通知に関するユースケースを表すインターフェース。
type NotificationService interface {
	// 通知一覧を取得する。
	ListNotifications(ctx context.Context, userID string, page, pageSize int) ([]*NotificationItem, int64, error)
	// 未読通知数を取得する。
	GetUnreadCount(ctx context.Context, userID string) (int64, error)
	// 指定通知を既読にする。
	MarkAsRead(ctx context.Context, notificationID, userID string) error
	// 全通知を既読にする。
	MarkAllAsRead(ctx context.Context, userID string) error
	// タグに映画が追加された通知を生成する。
	NotifyTagMovieAdded(ctx context.Context, tagID, tagMovieID, actorUserID string) error
	// タグがフォローされた通知を生成する。
	NotifyTagFollowed(ctx context.Context, tagID, actorUserID string) error
	// ユーザーがフォローされた通知を生成する。
	NotifyUserFollowed(ctx context.Context, followeeUserID, actorUserID string) error
	// フォロー中ユーザーが新しいタグを作成した通知を生成する。
	NotifyFollowingUserCreatedTag(ctx context.Context, tagID, actorUserID string) error
}

type notificationService struct {
	logger           *slog.Logger
	notifRepo        repository.NotificationRepository
	tagRepo          repository.TagRepository
	tagFollowerRepo  repository.TagFollowerRepository
	userFollowerRepo repository.UserFollowerRepository
}

// NotificationService を生成する。
func NewNotificationService(
	logger *slog.Logger,
	notifRepo repository.NotificationRepository,
	tagRepo repository.TagRepository,
	tagFollowerRepo repository.TagFollowerRepository,
	userFollowerRepo repository.UserFollowerRepository,
) NotificationService {
	return &notificationService{
		logger:           logger,
		notifRepo:        notifRepo,
		tagRepo:          tagRepo,
		tagFollowerRepo:  tagFollowerRepo,
		userFollowerRepo: userFollowerRepo,
	}
}

// 通知一覧を取得する。
func (s *notificationService) ListNotifications(ctx context.Context, userID string, page, pageSize int) ([]*NotificationItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}

	rows, total, err := s.notifRepo.ListByRecipient(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*NotificationItem{}, 0, nil
	}

	items := make([]*NotificationItem, 0, len(rows))
	for _, r := range rows {
		item := &NotificationItem{
			ID:               r.ID,
			NotificationType: r.NotificationType,
			IsRead:           r.IsRead,
			CreatedAt:        r.CreatedAt,
			MovieTitle:       r.MovieTitle,
		}

		// actor
		if r.ActorUserID != nil && r.ActorDisplayID != nil && r.ActorDisplayName != nil {
			item.Actor = &ActorSummary{
				ID:          *r.ActorUserID,
				DisplayID:   *r.ActorDisplayID,
				DisplayName: *r.ActorDisplayName,
				AvatarURL:   r.ActorAvatarURL,
			}
		}

		// tag
		if r.TagID != nil && r.TagTitle != nil {
			item.Tag = &TagSummaryForNotification{
				ID:    *r.TagID,
				Title: *r.TagTitle,
			}
		}

		items = append(items, item)
	}

	return items, total, nil
}

// 未読通知数を取得する。
func (s *notificationService) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	return s.notifRepo.CountUnread(ctx, userID)
}

// 指定通知を既読にする。
func (s *notificationService) MarkAsRead(ctx context.Context, notificationID, userID string) error {
	err := s.notifRepo.MarkAsRead(ctx, notificationID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotificationNotFound
		}
		return err
	}
	return nil
}

// 全通知を既読にする。
func (s *notificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.notifRepo.MarkAllAsRead(ctx, userID)
}

// タグに映画が追加された通知を生成する。
// 通知先: タグオーナー + タグフォロワー - アクター自身
func (s *notificationService) NotifyTagMovieAdded(ctx context.Context, tagID, tagMovieID, actorUserID string) error {
	tag, err := s.tagRepo.FindByID(ctx, tagID)
	if err != nil {
		return err
	}

	followerIDs, err := s.tagFollowerRepo.ListFollowerIDs(ctx, tagID)
	if err != nil {
		return err
	}

	// 通知先を集約: タグオーナー + フォロワー - アクター自身
	recipientSet := make(map[string]struct{})
	recipientSet[tag.UserID] = struct{}{}
	for _, id := range followerIDs {
		recipientSet[id] = struct{}{}
	}
	delete(recipientSet, actorUserID)

	if len(recipientSet) == 0 {
		return nil
	}

	notifications := make([]*model.Notification, 0, len(recipientSet))
	for recipientID := range recipientSet {
		actor := actorUserID
		tid := tagID
		tmid := tagMovieID
		notifications = append(notifications, &model.Notification{
			RecipientUserID:  recipientID,
			ActorUserID:      &actor,
			NotificationType: model.NotificationTypeTagMovieAdded,
			TagID:            &tid,
			TagMovieID:       &tmid,
		})
	}

	return s.notifRepo.CreateBatch(ctx, notifications)
}

// タグがフォローされた通知を生成する。
// 通知先: タグオーナー - アクター自身
func (s *notificationService) NotifyTagFollowed(ctx context.Context, tagID, actorUserID string) error {
	tag, err := s.tagRepo.FindByID(ctx, tagID)
	if err != nil {
		return err
	}

	// タグオーナーが自分自身の場合は通知しない
	if tag.UserID == actorUserID {
		return nil
	}

	actor := actorUserID
	tid := tagID
	notification := &model.Notification{
		RecipientUserID:  tag.UserID,
		ActorUserID:      &actor,
		NotificationType: model.NotificationTypeTagFollowed,
		TagID:            &tid,
	}

	return s.notifRepo.Create(ctx, notification)
}

// ユーザーがフォローされた通知を生成する。
// 通知先: フォローされたユーザー - アクター自身
func (s *notificationService) NotifyUserFollowed(ctx context.Context, followeeUserID, actorUserID string) error {
	// フォロー先が自分自身の場合は通知しない（通常ありえないが安全のため）
	if followeeUserID == actorUserID {
		return nil
	}

	actor := actorUserID
	notification := &model.Notification{
		RecipientUserID:  followeeUserID,
		ActorUserID:      &actor,
		NotificationType: model.NotificationTypeUserFollowed,
	}

	return s.notifRepo.Create(ctx, notification)
}

// フォロー中ユーザーが新しいタグを作成した通知を生成する。
// 通知先: アクターのフォロワー全員（公開タグのみ）
func (s *notificationService) NotifyFollowingUserCreatedTag(ctx context.Context, tagID, actorUserID string) error {
	// タグが公開かチェック
	tag, err := s.tagRepo.FindByID(ctx, tagID)
	if err != nil {
		return err
	}
	if !tag.IsPublic {
		return nil
	}

	followerIDs, err := s.userFollowerRepo.ListFollowerIDs(ctx, actorUserID)
	if err != nil {
		return err
	}

	if len(followerIDs) == 0 {
		return nil
	}

	notifications := make([]*model.Notification, 0, len(followerIDs))
	for _, recipientID := range followerIDs {
		actor := actorUserID
		tid := tagID
		notifications = append(notifications, &model.Notification{
			RecipientUserID:  recipientID,
			ActorUserID:      &actor,
			NotificationType: model.NotificationTypeFollowingUserCreatedTag,
			TagID:            &tid,
		})
	}

	return s.notifRepo.CreateBatch(ctx, notifications)
}

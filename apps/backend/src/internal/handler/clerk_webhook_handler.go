package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"cinetag-backend/src/internal/middleware"
	"cinetag-backend/src/internal/service"

	"github.com/gin-gonic/gin"
)

// Clerk Webhook の共通ペイロードを表します。
// data の中身はイベントタイプにより異なるため RawMessage で受け取ります。
// https://clerk.com/docs/guides/development/webhooks/overview#payload-structure
type clerkWebhookEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Clerk の user.created Webhook の data 部分を表します。
type clerkUserCreatedData struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	ImageURL       string `json:"image_url"`
	EmailAddresses []struct {
		EmailAddress string `json:"email_address"`
	} `json:"email_addresses"`
}

// Clerk の user.updated Webhook の data 部分を表します。
// user.created と同じ形式です。
type clerkUserUpdatedData struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	ImageURL       string `json:"image_url"`
	EmailAddresses []struct {
		EmailAddress string `json:"email_address"`
	} `json:"email_addresses"`
}

// Clerk の user.deleted Webhook の data 部分を表します。
type clerkUserDeletedData struct {
	ID string `json:"id"`
}

// Clerk Webhook を処理するハンドラーです。
type ClerkWebhookHandler struct {
	logger      *slog.Logger
	userService service.UserService
}

// ClerkWebhookHandler を初期化して返します。
func NewClerkWebhookHandler(logger *slog.Logger, userService service.UserService) *ClerkWebhookHandler {
	return &ClerkWebhookHandler{
		logger:      logger,
		userService: userService,
	}
}

// POST /api/v1/clerk/webhook を処理します。
// 現時点では svix 署名検証ロジックは未実装です。
// TODO: svix の署名検証を追加し、Clerk からの正当なリクエストのみを受け付ける。
func (h *ClerkWebhookHandler) HandleWebhook(c *gin.Context) {

	requestID := middleware.GetRequestID(c)

	// ペイロードをバインド
	var event clerkWebhookEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid webhook payload",
		})
		return
	}

	// 開始ログ（INFO）
	h.logger.Info("handler.HandleWebhook started",
		slog.String("request_id", requestID),
		slog.String("event_type", event.Type),
	)

	switch event.Type {
	case "user.created":
		var data clerkUserCreatedData
		if err := json.Unmarshal(event.Data, &data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid webhook data",
			})
			return
		}

		// デバッグログ（DEBUG）
		h.logger.Debug("handler.HandleWebhook user.created",
			slog.String("request_id", requestID),
			slog.String("clerk_user_id", data.ID),
			slog.String("username", data.Username),
			slog.String("first_name", data.FirstName),
			slog.String("last_name", data.LastName),
			slog.String("image_url", data.ImageURL),
		)

		email := ""
		if len(data.EmailAddresses) > 0 {
			// TODO: 複数のメールアドレスを無効化できる？
			email = data.EmailAddresses[0].EmailAddress
		}

		// ImageURL をオプショナルな *string 型に変換する。
		// 空文字列の場合は nil を設定し、値がある場合はローカル変数にコピーしてから
		// そのアドレスを取得する。これにより、構造体フィールドへの直接ポインタ取得を
		// 避け、エスケープ解析の最適化と独立性を確保する。
		var avatarURL *string
		if data.ImageURL != "" {
			url := data.ImageURL
			avatarURL = &url
		}

		// clerkUserInfo を作成
		clerkUser, err := service.NewClerkUserInfoFromWebhook(
			data.ID,
			email,
			data.FirstName,
			data.LastName,
			avatarURL,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to sync user",
			})
			return
		}

		if _, err := h.userService.EnsureUser(c.Request.Context(), clerkUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to sync user",
			})
			return
		}

		c.Status(http.StatusOK)
		return

	case "user.updated":
		var data clerkUserUpdatedData
		if err := json.Unmarshal(event.Data, &data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid webhook data",
			})
			return
		}

		// デバッグログ（DEBUG）
		h.logger.Debug("handler.HandleWebhook user.updated",
			slog.String("request_id", requestID),
			slog.String("clerk_user_id", data.ID),
			slog.String("image_url", data.ImageURL),
		)

		// ユーザーが存在しない場合は無視
		u, err := h.userService.FindUserByClerkUserID(c.Request.Context(), data.ID)
		if err != nil {
			if err == service.ErrUserNotFound {
				h.logger.Warn("handler.HandleWebhook user.updated: user not found",
					slog.String("request_id", requestID),
					slog.String("clerk_user_id", data.ID),
				)
				c.Status(http.StatusOK)
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to resolve user by clerk user id",
			})
			return
		}

		// avatar_url を更新
		var avatarURL *string
		if data.ImageURL != "" {
			url := data.ImageURL
			avatarURL = &url
		}

		if err := h.userService.UpdateUserFromClerk(c.Request.Context(), u.ID, avatarURL); err != nil {
			h.logger.Error("handler.HandleWebhook user.updated: failed to update user",
				slog.String("request_id", requestID),
				slog.String("user_id", u.ID),
				slog.String("error", err.Error()),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to update user",
			})
			return
		}

		c.Status(http.StatusOK)
		return

	case "user.deleted":
		var data clerkUserDeletedData
		if err := json.Unmarshal(event.Data, &data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid webhook data",
			})
			return
		}

		// デバッグログ（DEBUG）
		h.logger.Debug("handler.HandleWebhook user.deleted",
			slog.String("request_id", requestID),
			slog.String("clerk_user_id", data.ID),
		)

		// ユーザーが存在しない場合は成功とする
		u, err := h.userService.FindUserByClerkUserID(c.Request.Context(), data.ID)
		if err != nil {
			if err == service.ErrUserNotFound {
				c.Status(http.StatusOK)
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to resolve user by clerk user id",
			})
			return
		}

		// ユーザー削除
		if err := h.userService.DeactivateUser(c.Request.Context(), u.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to deactivate user",
			})
			return
		}

		c.Status(http.StatusOK)
		return

	default:
		// 他のイベントタイプは無視
		c.Status(http.StatusOK)
		return
	}
}

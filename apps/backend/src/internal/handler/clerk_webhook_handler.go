package handler

import (
	"net/http"

	"cinetag-backend/src/internal/service"

	"github.com/gin-gonic/gin"
)

// clerkUserCreatedEvent は Clerk の user.created Webhook ペイロードの一部を表します。
type clerkUserCreatedEvent struct {
	Type string `json:"type"`
	Data struct {
		ID             string `json:"id"`
		Username       string `json:"username"`
		FirstName      string `json:"first_name"`
		LastName       string `json:"last_name"`
		ImageURL       string `json:"image_url"`
		EmailAddresses []struct {
			EmailAddress string `json:"email_address"`
		} `json:"email_addresses"`
	} `json:"data"`
}

// ClerkWebhookHandler は Clerk Webhook を処理するハンドラーです。
type ClerkWebhookHandler struct {
	userService service.UserService
}

// NewClerkWebhookHandler は ClerkWebhookHandler を初期化して返します。
func NewClerkWebhookHandler(userService service.UserService) *ClerkWebhookHandler {
	return &ClerkWebhookHandler{
		userService: userService,
	}
}

// HandleWebhook は POST /api/v1/clerk/webhook を処理します。
//
// 現時点では svix 署名検証ロジックは未実装であり、payload の user.created イベントのみを処理します。
// TODO: svix の署名検証を追加し、Clerk からの正当なリクエストのみを受け付ける。
func (h *ClerkWebhookHandler) HandleWebhook(c *gin.Context) {

	var event clerkUserCreatedEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid webhook payload",
		})
		return
	}

	// 他のイベントタイプは無視
	if event.Type != "user.created" {
		c.Status(http.StatusOK)
		return
	}

	// TODO: user.updated, user.deletedイベントを考慮する必要はあるか。なんの時に発生するイベントか。

	email := ""
	if len(event.Data.EmailAddresses) > 0 {
		email = event.Data.EmailAddresses[0].EmailAddress
	}

	displayName := event.Data.Username
	if event.Data.FirstName != "" || event.Data.LastName != "" {
		displayName = event.Data.FirstName + " " + event.Data.LastName
	}

	var avatarURL *string
	if event.Data.ImageURL != "" {
		url := event.Data.ImageURL
		avatarURL = &url
	}

	clerkUser := service.ClerkUserInfo{
		ID:          event.Data.ID,
		Username:    event.Data.Username,
		Email:       email,
		DisplayName: displayName,
		AvatarURL:   avatarURL,
	}

	if _, err := h.userService.EnsureUser(c.Request.Context(), clerkUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to sync user",
		})
		return
	}

	c.Status(http.StatusOK)
}

package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"cinetag-backend/src/internal/service"

	"github.com/gin-gonic/gin"
)

// clerkWebhookEvent は Clerk Webhook の共通ペイロードを表します。
// data の中身はイベントタイプにより異なるため RawMessage で受け取ります。
type clerkWebhookEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// clerkUserCreatedData は Clerk の user.created Webhook の data 部分を表します。
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

// clerkUserDeletedData は Clerk の user.deleted Webhook の data 部分を表します。
type clerkUserDeletedData struct {
	ID string `json:"id"`
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
// 現時点では svix 署名検証ロジックは未実装です。
// TODO: svix の署名検証を追加し、Clerk からの正当なリクエストのみを受け付ける。
func (h *ClerkWebhookHandler) HandleWebhook(c *gin.Context) {

	// ペイロードをバインド
	var event clerkWebhookEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid webhook payload",
		})
		return
	}
	fmt.Println("[HandleWebhook] clerkWebhookEvent.Type", event.Type)

	switch event.Type {
	case "user.created":
		var data clerkUserCreatedData
		if err := json.Unmarshal(event.Data, &data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid webhook data",
			})
			return
		}

		fmt.Println("[HandleWebhook] clerkUserCreatedData.ID", data.ID)
		fmt.Println("[HandleWebhook] clerkUserCreatedData.Username", data.Username)
		fmt.Println("[HandleWebhook] clerkUserCreatedData.FirstName", data.FirstName)
		fmt.Println("[HandleWebhook] clerkUserCreatedData.LastName", data.LastName)
		fmt.Println("[HandleWebhook] clerkUserCreatedData.ImageURL", data.ImageURL)

		email := ""
		if len(data.EmailAddresses) > 0 {
			email = data.EmailAddresses[0].EmailAddress
		}

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

	case "user.deleted":
		var data clerkUserDeletedData
		if err := json.Unmarshal(event.Data, &data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid webhook data",
			})
			return
		}
		fmt.Println("[HandleWebhook] clerkUserDeletedData.ID", data.ID)

		// Clerk側で削除されたユーザーを、DB側で論理削除＋匿名化し、関連データをクリーンアップする。
		if err := h.userService.HandleClerkUserDeleted(c.Request.Context(), data.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to handle user deleted",
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

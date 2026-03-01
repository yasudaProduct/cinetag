package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/service"

	"github.com/gin-gonic/gin"
)

// 通知関連の HTTP ハンドラー。
type NotificationHandler struct {
	logger              *slog.Logger
	notificationService service.NotificationService
}

// NotificationHandler を初期化して返す。
func NewNotificationHandler(logger *slog.Logger, notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		logger:              logger,
		notificationService: notificationService,
	}
}

// コンテキストから認証済みユーザーを取得するヘルパー。
func getUserFromContext(c *gin.Context) *model.User {
	userRaw, exists := c.Get("user")
	if !exists {
		return nil
	}
	user, ok := userRaw.(*model.User)
	if !ok || user == nil {
		return nil
	}
	return user
}

// ListNotifications は通知一覧を取得する。
// GET /api/v1/notifications
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page := parseIntDefaultNotif(c.Query("page"), 1)
	pageSize := parseIntDefaultNotif(c.Query("page_size"), 20)
	if pageSize > 50 {
		pageSize = 50
	}
	unreadOnly := c.Query("unread_only") == "true"

	items, total, err := h.notificationService.ListNotifications(c.Request.Context(), user.ID, page, pageSize, unreadOnly)
	if err != nil {
		h.logger.Error("handler.ListNotifications failed",
			slog.Any("error", err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": items,
		"total":         total,
		"page":          page,
		"page_size":     pageSize,
	})
}

// GetUnreadCount は未読通知数を取得する。
// GET /api/v1/notifications/unread-count
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	count, err := h.notificationService.GetUnreadCount(c.Request.Context(), user.ID)
	if err != nil {
		h.logger.Error("handler.GetUnreadCount failed",
			slog.Any("error", err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unread_count": count,
	})
}

// MarkAsRead は指定の通知を既読にする。
// PATCH /api/v1/notifications/:notificationId/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notificationID := c.Param("notificationId")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification_id is required"})
		return
	}

	err := h.notificationService.MarkAsRead(c.Request.Context(), notificationID, user.ID)
	if err != nil {
		if errors.Is(err, service.ErrNotificationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
			return
		}
		h.logger.Error("handler.MarkAsRead failed",
			slog.Any("error", err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark as read"})
		return
	}

	c.Status(http.StatusNoContent)
}

// MarkAllAsRead は全通知を既読にする。
// PATCH /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err := h.notificationService.MarkAllAsRead(c.Request.Context(), user.ID)
	if err != nil {
		h.logger.Error("handler.MarkAllAsRead failed",
			slog.Any("error", err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark all as read"})
		return
	}

	c.Status(http.StatusNoContent)
}

func parseIntDefaultNotif(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

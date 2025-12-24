package handler

import (
	"errors"
	"net/http"
	"strconv"

	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/service"

	"github.com/gin-gonic/gin"
)

// UserHandler はユーザー関連のHTTPハンドラーを提供します。
type UserHandler struct {
	userService service.UserService
	tagService  service.TagService
}

// NewUserHandler は UserHandler を生成します。
func NewUserHandler(userService service.UserService, tagService service.TagService) *UserHandler {
	return &UserHandler{
		userService: userService,
		tagService:  tagService,
	}
}

// UserProfileResponse はユーザープロフィールのレスポンス形式です。
type UserProfileResponse struct {
	ID          string  `json:"id"`
	DisplayID   string  `json:"display_id"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	Bio         *string `json:"bio,omitempty"`
}

// GetMe は認証済みユーザー自身の情報を返します。
// GET /api/v1/users/me
func (h *UserHandler) GetMe(c *gin.Context) {
	userRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, ok := userRaw.(*model.User)
	if !ok || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}

	c.JSON(http.StatusOK, UserProfileResponse{
		ID:          user.ID,
		DisplayID:   user.DisplayID,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Bio:         user.Bio,
	})
}

// GetUserByDisplayID は display_id からユーザー情報を取得します。
// GET /api/v1/users/:displayId
func (h *UserHandler) GetUserByDisplayID(c *gin.Context) {
	displayID := c.Param("displayId")
	if displayID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "display_id is required"})
		return
	}

	user, err := h.userService.GetUserByDisplayID(c.Request.Context(), displayID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	c.JSON(http.StatusOK, UserProfileResponse{
		ID:          user.ID,
		DisplayID:   user.DisplayID,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Bio:         user.Bio,
	})
}

// ListUserTags はユーザーが作成したタグ一覧を取得します。
// GET /api/v1/users/:displayId/tags
func (h *UserHandler) ListUserTags(c *gin.Context) {
	displayID := c.Param("displayId")
	if displayID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "display_id is required"})
		return
	}

	// display_id からユーザーを取得
	user, err := h.userService.GetUserByDisplayID(c.Request.Context(), displayID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	page := parseIntDefaultUser(c.Query("page"), 1)
	pageSize := parseIntDefaultUser(c.Query("page_size"), 20)

	// 閲覧者が本人かどうかを判定
	publicOnly := true
	if viewerRaw, exists := c.Get("user"); exists {
		if viewer, ok := viewerRaw.(*model.User); ok && viewer != nil && viewer.ID == user.ID {
			publicOnly = false // 本人なら非公開タグも表示
		}
	}

	items, total, err := h.tagService.ListTagsByUserID(c.Request.Context(), user.ID, publicOnly, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list user tags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":       items,
		"page":        page,
		"page_size":   pageSize,
		"total_count": total,
	})
}

func parseIntDefaultUser(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

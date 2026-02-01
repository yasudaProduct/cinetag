package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"cinetag-backend/src/internal/middleware"
	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/service"

	"github.com/gin-gonic/gin"
)

// ユーザー関連のHTTPハンドラー。
type UserHandler struct {
	logger      *slog.Logger
	userService service.UserService
	tagService  service.TagService
}

// UserHandler を生成する。
func NewUserHandler(logger *slog.Logger, userService service.UserService, tagService service.TagService) *UserHandler {
	return &UserHandler{
		logger:      logger,
		userService: userService,
		tagService:  tagService,
	}
}

// ユーザープロフィールのレスポンス形式。
type UserProfileResponse struct {
	ID          string  `json:"id"`
	DisplayID   string  `json:"display_id"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	Bio         *string `json:"bio,omitempty"`
}

// ユーザー更新リクエストの形式。
type UpdateMeRequest struct {
	DisplayName *string `json:"display_name"`
}

// 認証済みユーザー自身の情報を返す。
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

// 認証済みユーザー自身の情報を更新する。
// PATCH /api/v1/users/me
func (h *UserHandler) UpdateMe(c *gin.Context) {
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

	var req UpdateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// 更新入力を構築
	input := service.UpdateUserInput{
		DisplayName: req.DisplayName,
	}

	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), user.ID, input)
	if err != nil {
		h.logger.Error("failed to update user",
			slog.String("user_id", user.ID),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, UserProfileResponse{
		ID:          updatedUser.ID,
		DisplayID:   updatedUser.DisplayID,
		DisplayName: updatedUser.DisplayName,
		AvatarURL:   updatedUser.AvatarURL,
		Bio:         updatedUser.Bio,
	})
}

// display_id からユーザー情報を取得する。
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

// ユーザーが作成したタグ一覧を取得する。
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

// FollowUser は指定ユーザーをフォローします。
// POST /api/v1/users/:displayId/follow
func (h *UserHandler) FollowUser(c *gin.Context) {
	displayID := c.Param("displayId")
	requestID := middleware.GetRequestID(c)

	// 開始ログ（INFO）
	attrs := []any{
		slog.String("request_id", requestID),
		slog.String("display_id", displayID),
	}

	// 認証済みの場合は user_id も含める
	if userVal, ok := c.Get("user"); ok {
		if user, ok2 := userVal.(*model.User); ok2 && user != nil {
			attrs = append(attrs, slog.String("user_id", user.ID))
		}
	}

	h.logger.Info("handler.FollowUser started", attrs...)
	if displayID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "display_id is required"})
		return
	}

	// 認証ユーザーを取得
	userRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	currentUser, ok := userRaw.(*model.User)
	if !ok || currentUser == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}

	// フォロー対象のユーザーを取得
	targetUser, err := h.userService.GetUserByDisplayID(c.Request.Context(), displayID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	// フォローを実行
	err = h.userService.FollowUser(c.Request.Context(), currentUser.ID, targetUser.ID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if errors.Is(err, service.ErrCannotFollowSelf) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot follow yourself"})
			return
		}
		if errors.Is(err, service.ErrAlreadyFollowing) {
			c.JSON(http.StatusConflict, gin.H{"error": "already following"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to follow user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully followed"})
}

// 指定ユーザーをアンフォローする。
// DELETE /api/v1/users/:displayId/follow
func (h *UserHandler) UnfollowUser(c *gin.Context) {
	displayID := c.Param("displayId")
	if displayID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "display_id is required"})
		return
	}

	// 認証ユーザーを取得
	userRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	currentUser, ok := userRaw.(*model.User)
	if !ok || currentUser == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}

	// アンフォロー対象のユーザーを取得
	targetUser, err := h.userService.GetUserByDisplayID(c.Request.Context(), displayID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	// アンフォローを実行
	err = h.userService.UnfollowUser(c.Request.Context(), currentUser.ID, targetUser.ID)
	if err != nil {
		if errors.Is(err, service.ErrNotFollowing) {
			c.JSON(http.StatusConflict, gin.H{"error": "not following"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unfollow user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully unfollowed"})
}

// 指定ユーザーがフォローしているユーザー一覧を取得する。
// GET /api/v1/users/:displayId/following
func (h *UserHandler) ListFollowing(c *gin.Context) {
	displayID := c.Param("displayId")
	if displayID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "display_id is required"})
		return
	}

	// ユーザーを取得
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

	users, total, err := h.userService.ListFollowing(c.Request.Context(), user.ID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list following"})
		return
	}

	items := make([]UserProfileResponse, len(users))
	for i, u := range users {
		items[i] = UserProfileResponse{
			ID:          u.ID,
			DisplayID:   u.DisplayID,
			DisplayName: u.DisplayName,
			AvatarURL:   u.AvatarURL,
			Bio:         u.Bio,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"items":       items,
		"page":        page,
		"page_size":   pageSize,
		"total_count": total,
	})
}

// 指定ユーザーをフォローしているユーザー一覧を取得する。
// GET /api/v1/users/:displayId/followers
func (h *UserHandler) ListFollowers(c *gin.Context) {
	displayID := c.Param("displayId")
	if displayID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "display_id is required"})
		return
	}

	// ユーザーを取得
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

	users, total, err := h.userService.ListFollowers(c.Request.Context(), user.ID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list followers"})
		return
	}

	items := make([]UserProfileResponse, len(users))
	for i, u := range users {
		items[i] = UserProfileResponse{
			ID:          u.ID,
			DisplayID:   u.DisplayID,
			DisplayName: u.DisplayName,
			AvatarURL:   u.AvatarURL,
			Bio:         u.Bio,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"items":       items,
		"page":        page,
		"page_size":   pageSize,
		"total_count": total,
	})
}

// ユーザーのフォロー数・フォロワー数を取得する。
// GET /api/v1/users/:displayId/follow-stats
func (h *UserHandler) GetUserFollowStats(c *gin.Context) {
	displayID := c.Param("displayId")
	if displayID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "display_id is required"})
		return
	}

	// ユーザーを取得
	user, err := h.userService.GetUserByDisplayID(c.Request.Context(), displayID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	following, followers, err := h.userService.GetFollowStats(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get follow stats"})
		return
	}

	// 認証ユーザーがこのユーザーをフォローしているか確認
	isFollowing := false
	if viewerRaw, exists := c.Get("user"); exists {
		if viewer, ok := viewerRaw.(*model.User); ok && viewer != nil {
			isFollowing, _ = h.userService.IsFollowing(c.Request.Context(), viewer.ID, user.ID)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"following_count": following,
		"followers_count": followers,
		"is_following":    isFollowing,
	})
}

// ユーザーのページ番号とページサイズを取得する。
func parseIntDefaultUser(s string, def int) int {
	// ページ番号が空の場合はデフォルト値を返す
	if s == "" {
		return def
	}

	// ページ番号を整数に変換。変換に失敗した場合はデフォルト値を返す。
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}

	return v
}

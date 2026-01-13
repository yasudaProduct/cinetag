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

// タグ関連の HTTP ハンドラー。
type TagHandler struct {
	logger     *slog.Logger
	tagService service.TagService
}

// TagHandler を初期化して返す。
func NewTagHandler(logger *slog.Logger, tagService service.TagService) *TagHandler {
	return &TagHandler{
		logger:     logger,
		tagService: tagService,
	}
}

// タグ作成リクエストボディの構造。
// user_id はクライアントからは受け取らず、AuthMiddleware によりコンテキストに設定された認証済みユーザー情報から取得する。
type createTagRequest struct {
	Title          string  `json:"title" binding:"required"`
	Description    *string `json:"description"`
	CoverImageURL  *string `json:"cover_image_url"`
	IsPublic       *bool   `json:"is_public"`
	AddMoviePolicy *string `json:"add_movie_policy"`
}

// タグに映画を追加するリクエストボディの構造。
type addTagMovieRequest struct {
	TmdbMovieID int     `json:"tmdb_movie_id" binding:"required"`
	Note        *string `json:"note"`
	Position    int     `json:"position"`
}

// タグのメタ情報を更新するリクエストボディの構造。
type updateTagRequest struct {
	Title          *string  `json:"title"`
	Description    **string `json:"description"`
	CoverImageURL  **string `json:"cover_image_url"`
	IsPublic       *bool    `json:"is_public"`
	AddMoviePolicy *string  `json:"add_movie_policy"`
}

// タグ詳細を取得する。
// GET /api/v1/tags/:tagId
func (h *TagHandler) GetTagDetail(c *gin.Context) {
	tagID := c.Param("tagId")

	var viewerUserID *string
	if userVal, ok := c.Get("user"); ok {
		if user, ok2 := userVal.(*model.User); ok2 && user != nil && user.ID != "" {
			id := user.ID
			viewerUserID = &id
		}
	}

	// タグ詳細を取得する。
	out, err := h.tagService.GetTagDetail(c.Request.Context(), tagID, viewerUserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTagNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		case errors.Is(err, service.ErrTagPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get tag detail"})
		}
		return
	}

	c.JSON(http.StatusOK, out)
}

// タグ内の映画一覧を取得する。
func (h *TagHandler) ListTagMovies(c *gin.Context) {
	tagID := c.Param("tagId")

	page := parseIntDefault(c.Query("page"), 1)
	pageSize := parseIntDefault(c.Query("page_size"), 50)

	var viewerUserID *string
	if userVal, ok := c.Get("user"); ok {
		if user, ok2 := userVal.(*model.User); ok2 && user != nil && user.ID != "" {
			id := user.ID
			viewerUserID = &id
		}
	}

	items, total, err := h.tagService.ListTagMovies(c.Request.Context(), tagID, viewerUserID, page, pageSize)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTagNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		case errors.Is(err, service.ErrTagPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tag movies"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":       items,
		"page":        page,
		"page_size":   pageSize,
		"total_count": total,
	})
}

// UpdateTag はタグのメタ情報を更新します（作成者のみ）。
func (h *TagHandler) UpdateTag(c *gin.Context) {
	tagID := c.Param("tagId")

	// AuthMiddleware によってコンテキストに設定されたユーザー情報を取得
	userVal, exists := c.Get("user")
	if !exists {
		// ユーザー情報がない場合、401 Unauthorizedを返す
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, ok := userVal.(*model.User)
	if !ok || user == nil {
		// ユーザー情報が無効な場合、500 Internal Server Errorを返す
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user in context"})
		return
	}

	var req updateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// リクエストボディが無効な場合、400 Bad Requestを返す
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// タグを更新する。
	out, err := h.tagService.UpdateTag(c.Request.Context(), tagID, user.ID, service.UpdateTagPatch{
		Title:          req.Title,
		Description:    req.Description,
		CoverImageURL:  req.CoverImageURL,
		IsPublic:       req.IsPublic,
		AddMoviePolicy: req.AddMoviePolicy,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTagNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		case errors.Is(err, service.ErrTagPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tag"})
		}
		return
	}

	c.JSON(http.StatusOK, out)
}

// @Summary 公開タグ一覧を取得
// @Description 公開タグ一覧を取得
// @Tags tags
// @Accept json
// @Produce json
// @Param q query string false "タイトル検索用キーワード"
// @Param sort query string false "popular / recent / movie_count"
// @Param page query int false "ページ番号"
// @Param page_size query int false "1ページあたり件数"
// @Success 200 {object}
// @Failure 500 {object}
// @Router /api/v1/tags [get]
func (h *TagHandler) ListPublicTags(c *gin.Context) {
	q := c.Query("q")
	sort := c.Query("sort")

	page := parseIntDefault(c.Query("page"), 1)
	pageSize := parseIntDefault(c.Query("page_size"), 20)

	items, total, err := h.tagService.ListPublicTags(c.Request.Context(), q, sort, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list tags",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":       items,
		"page":        page,
		"page_size":   pageSize,
		"total_count": total,
	})
}

// @Summary タグを作成
// @Description タグを作成
// @Tags tags
// @Accept json
// @Produce json
// @Param request body createTagRequest true "タグ作成リクエスト"
// @Success 201 {object}
// @Failure 400 {object}
// @Failure 500 {object}
// @Router api/v1/tags [post]
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req createTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// AuthMiddleware によってコンテキストに設定されたユーザー情報を取得
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}
	user, ok := userVal.(*model.User)
	if !ok || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user in context",
		})
		return
	}

	// シンプルなバリデーション（タイトル長・説明長）
	if l := len([]rune(req.Title)); l == 0 || l > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "title must be between 1 and 100 characters",
		})
		return
	}
	if req.Description != nil {
		if l := len([]rune(*req.Description)); l > 500 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "description must be 500 characters or less",
			})
			return
		}
	}

	tag, err := h.tagService.CreateTag(c.Request.Context(), service.CreateTagInput{
		UserID:         user.ID,
		Title:          req.Title,
		Description:    req.Description,
		CoverImageURL:  req.CoverImageURL,
		IsPublic:       req.IsPublic,
		AddMoviePolicy: req.AddMoviePolicy,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create tag",
		})
		return
	}

	// レスポンスは api-spec に合わせて必要なフィールドのみ返す
	c.JSON(http.StatusCreated, gin.H{
		"id":               tag.ID,
		"title":            tag.Title,
		"description":      tag.Description,
		"cover_image_url":  tag.CoverImageURL,
		"is_public":        tag.IsPublic,
		"add_movie_policy": tag.AddMoviePolicy,
		"movie_count":      tag.MovieCount,
		"follower_count":   tag.FollowerCount,
		"created_at":       tag.CreatedAt,
		"updated_at":       tag.UpdatedAt,
	})
}

// @name AddMovieToTag
// @Summary タグに映画を追加
// @Description タグに映画を追加
// @Tags tags
// @Accept json
// @Produce json
// @Param tagId path string true "タグID"
// @Param request body addTagMovieRequest true "映画追加リクエスト"
func (h *TagHandler) AddMovieToTag(c *gin.Context) {
	tagID := c.Param("tagId")
	requestID := middleware.GetRequestID(c)

	// 開始ログ（INFO）
	attrs := []any{
		slog.String("request_id", requestID),
		slog.String("tag_id", tagID),
	}

	// 認証済みの場合は user_id も含める
	if userVal, ok := c.Get("user"); ok {
		if user, ok2 := userVal.(*model.User); ok2 && user != nil {
			attrs = append(attrs, slog.String("user_id", user.ID))
		}
	}

	h.logger.Info("handler.AddMovieToTag started", attrs...)

	var req addTagMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}
	if req.TmdbMovieID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "tmdb_movie_id must be a positive integer",
		})
		return
	}
	if req.Position < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "position must be 0 or greater",
		})
		return
	}

	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}
	user, ok := userVal.(*model.User)
	if !ok || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user in context",
		})
		return
	}

	// タグに映画を追加する。
	tagMovie, err := h.tagService.AddMovieToTag(c.Request.Context(), service.AddMovieToTagInput{
		TagID:       tagID,
		UserID:      user.ID,
		TmdbMovieID: req.TmdbMovieID,
		Note:        req.Note,
		Position:    req.Position,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTagNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		case errors.Is(err, service.ErrTagPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, service.ErrTagMovieAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "movie already added to tag"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add movie to tag"})
		}
		return
	}

	c.JSON(http.StatusCreated, tagMovie)
}

// RemoveMovieFromTag はタグから映画を削除します。
func (h *TagHandler) RemoveMovieFromTag(c *gin.Context) {
	tagMovieID := c.Param("tagMovieId")

	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}
	user, ok := userVal.(*model.User)
	if !ok || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user in context",
		})
		return
	}

	err := h.tagService.RemoveMovieFromTag(c.Request.Context(), tagMovieID, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTagMovieNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "tag movie not found"})
		case errors.Is(err, service.ErrTagNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		case errors.Is(err, service.ErrTagPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove movie from tag"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// FollowTag はタグをフォローします。
// POST /api/v1/tags/:tagId/follow
func (h *TagHandler) FollowTag(c *gin.Context) {
	tagID := c.Param("tagId")

	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, ok := userVal.(*model.User)
	if !ok || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user in context"})
		return
	}

	err := h.tagService.FollowTag(c.Request.Context(), tagID, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTagNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		case errors.Is(err, service.ErrTagPermissionDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, service.ErrAlreadyFollowingTag):
			c.JSON(http.StatusConflict, gin.H{"error": "already following"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to follow tag"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully followed"})
}

// UnfollowTag はタグのフォローを解除します。
// DELETE /api/v1/tags/:tagId/follow
func (h *TagHandler) UnfollowTag(c *gin.Context) {
	tagID := c.Param("tagId")

	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, ok := userVal.(*model.User)
	if !ok || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user in context"})
		return
	}

	err := h.tagService.UnfollowTag(c.Request.Context(), tagID, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTagNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		case errors.Is(err, service.ErrNotFollowingTag):
			c.JSON(http.StatusConflict, gin.H{"error": "not following"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unfollow tag"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully unfollowed"})
}

// GetTagFollowStatus はユーザーがタグをフォローしているかチェックします。
// GET /api/v1/tags/:tagId/follow-status
func (h *TagHandler) GetTagFollowStatus(c *gin.Context) {
	tagID := c.Param("tagId")

	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, ok := userVal.(*model.User)
	if !ok || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user in context"})
		return
	}

	isFollowing, err := h.tagService.IsFollowingTag(c.Request.Context(), tagID, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get follow status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_following": isFollowing})
}

// ListTagFollowers はタグのフォロワー一覧を取得します。
// GET /api/v1/tags/:tagId/followers
func (h *TagHandler) ListTagFollowers(c *gin.Context) {
	tagID := c.Param("tagId")

	page := parseIntDefault(c.Query("page"), 1)
	pageSize := parseIntDefault(c.Query("page_size"), 20)

	users, total, err := h.tagService.ListTagFollowers(c.Request.Context(), tagID, page, pageSize)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTagNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tag followers"})
		}
		return
	}

	// ユーザー情報をレスポンス用に変換
	items := make([]gin.H, 0, len(users))
	for _, u := range users {
		items = append(items, gin.H{
			"id":           u.ID,
			"display_id":   u.DisplayID,
			"display_name": u.DisplayName,
			"avatar_url":   u.AvatarURL,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items":       items,
		"page":        page,
		"page_size":   pageSize,
		"total_count": total,
	})
}

// ListFollowingTags はログインユーザーがフォローしているタグ一覧を取得します。
// GET /api/v1/me/following-tags
func (h *TagHandler) ListFollowingTags(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, ok := userVal.(*model.User)
	if !ok || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user in context"})
		return
	}

	page := parseIntDefault(c.Query("page"), 1)
	pageSize := parseIntDefault(c.Query("page_size"), 20)

	items, total, err := h.tagService.ListFollowingTags(c.Request.Context(), user.ID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list following tags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":       items,
		"page":        page,
		"page_size":   pageSize,
		"total_count": total,
	})
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

package handler

import (
	"net/http"
	"strconv"

	"cinetag-backend/internal/model"
	"cinetag-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// TagHandler はタグ関連の HTTP ハンドラーを提供します。
type TagHandler struct {
	tagService service.TagService
}

// NewTagHandler は TagHandler を初期化して返します。
func NewTagHandler(tagService service.TagService) *TagHandler {
	return &TagHandler{
		tagService: tagService,
	}
}

// createTagRequest はタグ作成リクエストボディの構造を表します。
// user_id はクライアントからは受け取らず、AuthMiddleware によりコンテキストに設定された
// 認証済みユーザー情報から取得します。
type createTagRequest struct {
	Title         string  `json:"title" binding:"required"`
	Description   *string `json:"description"`
	CoverImageURL *string `json:"cover_image_url"`
	IsPublic      *bool   `json:"is_public"`
}

// ListPublicTags は GET /api/v1/tags を処理し、公開タグ一覧を返します。
//
// クエリパラメータ:
//   - q:     タイトル検索用キーワード（任意）
//   - sort:  popular / recent / movie_count（任意）
//   - page:  ページ番号（デフォルト 1）
//   - page_size: 1ページあたり件数（デフォルト 20, 最大 100）
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

// CreateTag は POST /api/v1/tags を処理し、新しいタグを作成します。
// user_id は AuthMiddleware によりコンテキストに設定された *model.User から取得します。
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
		UserID:        user.ID,
		Title:         req.Title,
		Description:   req.Description,
		CoverImageURL: req.CoverImageURL,
		IsPublic:      req.IsPublic,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create tag",
		})
		return
	}

	// レスポンスは api-spec に合わせて必要なフィールドのみ返す
	c.JSON(http.StatusCreated, gin.H{
		"id":              tag.ID,
		"title":           tag.Title,
		"description":     tag.Description,
		"cover_image_url": tag.CoverImageURL,
		"is_public":       tag.IsPublic,
		"movie_count":     tag.MovieCount,
		"follower_count":  tag.FollowerCount,
		"created_at":      tag.CreatedAt,
		"updated_at":      tag.UpdatedAt,
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

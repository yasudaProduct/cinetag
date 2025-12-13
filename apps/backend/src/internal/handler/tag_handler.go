package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/service"

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

type addTagMovieRequest struct {
	TmdbMovieID int     `json:"tmdb_movie_id" binding:"required"`
	Note        *string `json:"note"`
	Position    int     `json:"position"`
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

// AddMovieToTag はタグに映画を追加します。
func (h *TagHandler) AddMovieToTag(c *gin.Context) {
	tagID := c.Param("tagId")
	fmt.Println("tagID", tagID)

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

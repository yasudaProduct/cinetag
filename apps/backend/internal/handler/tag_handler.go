package handler

import (
	"net/http"

	"cinetag-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// TagHandler は映画タグ関連の HTTP ハンドラーを提供します。
type TagHandler struct {
	service service.TagService
}

func NewTagHandler(s service.TagService) *TagHandler {
	return &TagHandler{
		service: s,
	}
}

// GET /api/v1/tags
func (h *TagHandler) ListTags(c *gin.Context) {
	tags, err := h.service.ListTags(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list tags",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tags": tags,
	})
}



package router

import (
	"net/http"

	// "cinetag-backend/internal/handler"
	// "cinetag-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	// 依存関係の組み立て（暫定的にここでモックサービスを生成）
	// tagService := service.NewMockTagService()
	// tagHandler := handler.NewTagHandler(tagService)

	// ヘルスチェック用エンドポイント
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// すべての API は /api/v1 配下にまとめる
	// api := r.Group("/api/v1")
	// {

	// }

	return r
}

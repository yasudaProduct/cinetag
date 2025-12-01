package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	// ヘルスチェック用エンドポイント
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// すべての API は /api/v1 配下にまとめる
	api := r.Group("/api/v1")
	{
		// 例: 映画カテゴリ関連 API を今後ここに追加
		_ = api
	}

	return r
}



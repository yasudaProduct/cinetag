package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// 実行ポートを環境変数から取得（デフォルト: 8080）
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Gin のデフォルトルーターを作成
	router := gin.Default()

	// ヘルスチェック用エンドポイント
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// 今後の API は /api/v1 以下に生やしていく想定
	api := router.Group("/api/v1")
	{
		// 例: 映画カテゴリ関連 API を今後ここに追加
		_ = api
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}



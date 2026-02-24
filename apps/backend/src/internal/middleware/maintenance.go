package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// NewMaintenanceMiddleware はメンテナンスモードを制御するミドルウェアを返します。
//
// 環境変数 MAINTENANCE_MODE が "true" (大文字小文字不問) の場合、
// /health を除く全てのリクエストに対して 503 Service Unavailable を返します。
//
// 用途: 破壊的マイグレーション実行中にアプリケーションへのリクエストをブロック
func NewMaintenanceMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isMaintenanceMode() {
			c.Next()
			return
		}

		// /health はヘルスチェック用に常に通す（Cloud Run の死活監視）
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		logger.Warn("maintenance mode: rejecting request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		)

		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"error":   "service_unavailable",
			"message": "The service is currently under maintenance. Please try again later.",
		})
	}
}

func isMaintenanceMode() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("MAINTENANCE_MODE")), "true")
}

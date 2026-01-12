package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// NewRecoveryMiddleware はパニックをリカバリーしてログ出力するミドルウェアを返します。
//
// gin.Recovery() の代替として使用し、パニック発生時に以下の情報をログ出力します:
//   - request_id: リクエストID（設定されている場合）
//   - method: HTTPメソッド
//   - path: リクエストパス
//   - error: パニックの内容
//   - stack: スタックトレース
func NewRecoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// スタックトレースを取得
				stack := string(debug.Stack())

				// リクエストIDを取得（設定されている場合）
				requestID := GetRequestID(c)

				// エラーログを出力
				logger.Error("panic recovered",
					slog.String("request_id", requestID),
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.Any("error", err),
					slog.String("stack", stack),
				)

				// 500 Internal Server Error を返す
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()

		c.Next()
	}
}

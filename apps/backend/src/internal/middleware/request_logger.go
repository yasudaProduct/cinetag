package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ContextKeyRequestID はコンテキストに格納するリクエストIDのキー。
const ContextKeyRequestID = "request_id"

// NewRequestLoggerMiddleware はリクエストログを出力するミドルウェアを返す。
//
// 各リクエストに UUID を付与し、リクエスト完了時に以下の情報をログ出力する:
//   - request_id: リクエストを識別するUUID
//   - method: HTTPメソッド
//   - path: リクエストパス
//   - status: HTTPステータスコード
//   - latency: 処理時間
//   - client_ip: クライアントIP
func NewRequestLoggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// リクエストIDを生成してコンテキストに設定
		requestID := uuid.NewString()
		c.Set(ContextKeyRequestID, requestID)

		// リクエスト開始時刻を記録
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 次のハンドラーを実行
		c.Next()

		// リクエスト完了後のログ出力
		latency := time.Since(start)
		status := c.Writer.Status()

		// クエリパラメータがある場合はパスに付与
		if query != "" {
			path = path + "?" + query
		}

		// ログ属性を構築
		attrs := []any{
			slog.String("request_id", requestID),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
		}

		// エラーがある場合は追加
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("errors", c.Errors.String()))
		}

		// ステータスコードに応じてログレベルを変更
		switch {
		case status >= 500:
			logger.Error("request", attrs...)
		case status >= 400:
			logger.Warn("request", attrs...)
		default:
			logger.Info("request", attrs...)
		}
	}
}

// GetRequestID はコンテキストからリクエストIDを取得します。
// リクエストIDが見つからない場合は空文字列を返します。
func GetRequestID(c *gin.Context) string {
	if id, ok := c.Get(ContextKeyRequestID); ok {
		if requestID, ok := id.(string); ok {
			return requestID
		}
	}
	return ""
}

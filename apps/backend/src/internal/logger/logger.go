package logger

import (
	"log/slog"
	"os"
	"strings"
)

// 環境変数に基づいて設定された *slog.Logger を返す。
//
// 環境変数:
//   - LOG_LEVEL: debug / info / warn / error (デフォルト: info)
//   - GIN_MODE: release (JSON形式) / development (テキスト形式)
func NewLogger() *slog.Logger {
	level := parseLogLevel(os.Getenv("LOG_LEVEL"))
	isProduction := os.Getenv("GIN_MODE") == "release"

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if isProduction {
		// 本番環境: JSON形式
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// 開発環境: テキスト形式
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// 文字列からログレベルを解析する。
// 不明な値の場合は slog.LevelInfo を返す。
func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

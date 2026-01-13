package testutil

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/gin-gonic/gin"
)

// NewTestRouter は Gin をテスト用設定で初期化して返します。
// テスト側で必要なルーティングを登録して使います。
func NewTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// NewTestLogger はテスト用のロガーを返します。
// 出力は os.Stderr に送られますが、通常はテスト中は表示されません。
func NewTestLogger() *slog.Logger {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	return slog.New(handler)
}

// PerformRequest は http.Handler に対して疑似リクエストを実行し、レスポンスを返します。
func PerformRequest(h http.Handler, method, path string, body []byte, headers map[string]string) *httptest.ResponseRecorder {
	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	return rw
}

package testutil

import (
	"net/http"

	"cinetag-backend/src/internal/model"

	"github.com/gin-gonic/gin"
)

// TestAuthMiddleware はテスト用の認証バイパスミドルウェアです。
// X-Test-User-ID ヘッダーの値で testUsers マップからユーザーを取得し、
// c.Set("user", ...) でコンテキストに設定します。
// ヘッダーが無い場合は 401 を返します。
func TestAuthMiddleware(testUsers map[string]*model.User) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-Test-User-ID")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		user, ok := testUsers[userID]
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "test user not found"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

// TestOptionalAuthMiddleware はテスト用のオプショナル認証バイパスミドルウェアです。
// X-Test-User-ID ヘッダーがあればユーザーをセットし、なければそのまま通過します。
func TestOptionalAuthMiddleware(testUsers map[string]*model.User) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-Test-User-ID")
		if userID == "" {
			c.Next()
			return
		}

		if user, ok := testUsers[userID]; ok {
			c.Set("user", user)
		}

		c.Next()
	}
}

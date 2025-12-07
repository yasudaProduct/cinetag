package middleware

import (
	"net/http"
	"strings"

	"cinetag-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// NewAuthMiddleware は、Clerk トークンの検証（暫定）と
// users テーブルとの同期を行う認証ミドルウェアを返します。
//
// NOTE:
//   - 現時点では Clerk の JWT 検証ロジックは未実装であり、
//     Authorization ヘッダの Bearer トークン値をそのまま clerk_user_id として扱います。
//   - 将来的には Clerk 公式 SDK / 公開鍵を利用してトークンを検証し、
//     ClerkUserInfo を組み立てる実装に差し替える想定です。
func NewAuthMiddleware(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}

		rawToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if rawToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}

		// TODO: ここで Clerk の JWT を検証し、クレームからユーザー情報を取り出す。
		// 現時点では簡易実装として、トークン値をそのまま clerk_user_id / username として扱う。
		clerkUser := service.ClerkUserInfo{
			ID:          rawToken,
			Username:    rawToken,
			DisplayName: rawToken,
			Email:       rawToken + "@example.com",
		}

		user, err := userService.EnsureUser(c.Request.Context(), clerkUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to ensure user",
			})
			c.Abort()
			return
		}

		// 後続のハンドラーから参照できるよう、コンテキストに格納する
		c.Set("user", user)

		c.Next()
	}
}

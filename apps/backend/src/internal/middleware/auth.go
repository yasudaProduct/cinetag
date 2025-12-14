package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

	"cinetag-backend/src/internal/service"

	"github.com/gin-gonic/gin"
)

// NewAuthMiddleware は、Clerk トークンの検証（暫定）と
// users テーブルとの同期を行う認証ミドルウェアを返します。
//
// NOTE:
//   - Clerk の JWKS で RS256 JWT を検証し、sub（Clerk user ID）を信頼できる形で取得します。
//   - JWKS の取得先は環境変数 `CLERK_JWKS_URL` に設定してください。
//   - 必要なら `CLERK_ISSUER` / `CLERK_AUDIENCE` も指定し、iss/aud の検証を有効化できます。
func NewAuthMiddleware(userService service.UserService) gin.HandlerFunc {
	jwksURL := os.Getenv("CLERK_JWKS_URL")
	issuer := os.Getenv("CLERK_ISSUER")
	audience := os.Getenv("CLERK_AUDIENCE")

	validator, err := NewClerkJWTValidator(jwksURL, issuer, audience)
	if err != nil {
		// ルーティング初期化時に気づけるようログに出し、リクエストは 500 を返す
		log.Printf("AuthMiddleware misconfigured: %v", err)
		validator = nil
	}

	return func(c *gin.Context) {
		if validator == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "auth middleware misconfigured",
			})
			c.Abort()
			return
		}

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

		claims, err := validator.Verify(c.Request.Context(), rawToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}

		sub, _ := claims["sub"].(string)
		sub = strings.TrimSpace(sub)
		if sub == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}

		email := ""
		if s, ok := claims["email"].(string); ok {
			email = strings.TrimSpace(s)
		}
		if email == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}

		userName := "未設定"
		if s, ok := claims["full_name"].(string); ok && strings.TrimSpace(s) != "" {
			userName = strings.TrimSpace(s)
		}

		displayName := userName
		if s, ok := claims["first_name"].(string); ok && strings.TrimSpace(s) != "" {
			displayName = strings.TrimSpace(s)
		}

		var imageURL *string
		if s, ok := claims["image_url"].(string); ok && strings.TrimSpace(s) != "" {
			url := strings.TrimSpace(s)
			imageURL = &url
		}

		clerkUser := service.ClerkUserInfo{
			ID:          sub,
			Username:    userName,
			DisplayName: displayName,
			Email:       email,
			AvatarURL:   imageURL,
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

package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

	"cinetag-backend/src/internal/service"

	"github.com/gin-gonic/gin"
)

// NewOptionalAuthMiddleware は Authorization ヘッダーが付いている場合のみ Clerk JWT を検証し、
// users テーブルと同期した User をコンテキストに設定します。
//
// - Authorization が無い場合: そのまま通す（匿名アクセス）
// - Authorization があるが不正: 401 を返す
//
// NOTE: 検証器の設定（CLERK_JWKS_URL 等）は AuthMiddleware と同様です。
func NewOptionalAuthMiddleware(userService service.UserService) gin.HandlerFunc {
	jwksURL := os.Getenv("CLERK_JWKS_URL")
	issuer := os.Getenv("CLERK_ISSUER")
	audience := os.Getenv("CLERK_AUDIENCE")

	validator, err := NewClerkJWTValidator(jwksURL, issuer, audience)
	if err != nil {
		log.Printf("OptionalAuthMiddleware misconfigured: %v", err)
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

		// Authorization Bearer トークンが付いている場合のみ検証する
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.Next()
			return
		}
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// Bearer トークンを取得する
		rawToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if rawToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// トークンを検証する
		claims, err := validator.Verify(c.Request.Context(), rawToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// ClerkUserInfo を作成する
		clerkUser, err := service.NewClerkUserInfoFromJWTClaims(claims)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		user, err := userService.EnsureUser(c.Request.Context(), clerkUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to ensure user"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

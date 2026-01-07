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

		rawToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if rawToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		claims, err := validator.Verify(c.Request.Context(), rawToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		sub, _ := claims["sub"].(string)
		sub = strings.TrimSpace(sub)
		if sub == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		email := ""
		if s, ok := claims["email"].(string); ok {
			email = strings.TrimSpace(s)
		}
		if email == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		firstName := ""
		if s, ok := claims["first_name"].(string); ok {
			firstName = strings.TrimSpace(s)
		}

		lastName := ""
		if s, ok := claims["last_name"].(string); ok {
			lastName = strings.TrimSpace(s)
		}

		var imageURL *string
		if s, ok := claims["image_url"].(string); ok && strings.TrimSpace(s) != "" {
			url := strings.TrimSpace(s)
			imageURL = &url
		}

		clerkUser := service.ClerkUserInfo{
			ID:        sub,
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
			AvatarURL: imageURL,
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

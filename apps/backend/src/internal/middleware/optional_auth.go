package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"cinetag-backend/src/internal/service"

	"github.com/gin-gonic/gin"
)

// Authorization ヘッダーが付いている場合のみ Clerk JWT を検証し、
// users テーブルと同期した User をコンテキストに設定する。
// - Authorization が無い場合: そのまま通す（匿名アクセス）
// - Authorization があるが不正: 401 を返す
// - CLERK_JWKS_URL 等の検証器の設定は AuthMiddleware と同様。
func NewOptionalAuthMiddleware(logger *slog.Logger, userService service.UserService) gin.HandlerFunc {
	// 初期化ログ（DEBUG）
	logger.Debug("middleware.NewOptionalAuthMiddleware initialized")

	jwksURL := os.Getenv("CLERK_JWKS_URL")
	issuer := os.Getenv("CLERK_ISSUER")
	audience := os.Getenv("CLERK_AUDIENCE")

	// Clerk JWT 検証器を生成する。
	validator, err := NewClerkJWTValidator(jwksURL, issuer, audience)
	if err != nil {
		logger.Error("OptionalAuthMiddleware misconfigured", slog.Any("error", err))
		validator = nil
	}

	return func(c *gin.Context) {

		// Clerk JWT 検証器が生成できなかった場合は 500 を返す。
		if validator == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "auth middleware misconfigured",
			})
			c.Abort()
			return
		}

		// Authorization ヘッダーを取得する。
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			// Authorization ヘッダーが空の場合はそのまま通す。
			c.Next()
			return
		}
		if !strings.HasPrefix(authHeader, "Bearer ") {
			// Authorization ヘッダーが Bearer 形式でない場合は 401 を返す。
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

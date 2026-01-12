package middleware

import (
	"fmt"
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
	fmt.Println("[NewAuthMiddleware] NewAuthMiddleware")
	jwksURL := os.Getenv("CLERK_JWKS_URL")
	issuer := os.Getenv("CLERK_ISSUER")
	audience := os.Getenv("CLERK_AUDIENCE")

	// Clerk JWT 検証器を生成する。
	validator, err := NewClerkJWTValidator(jwksURL, issuer, audience)
	if err != nil {
		// ルーティング初期化時に気づけるようログに出し、リクエストは 500 を返す
		log.Printf("AuthMiddleware misconfigured: %v", err)
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
		authHeader := c.GetHeader("Authorization")
		// Authorization ヘッダーが空か Bearer 形式でない場合は 401 を返す。
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}

		// Bearer トークンを取得する。
		rawToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if rawToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}

		// Bearer トークンを検証する。
		claims, err := validator.Verify(c.Request.Context(), rawToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}

		// ClerkUserInfo を作成する。
		clerkUser, err := service.NewClerkUserInfoFromJWTClaims(claims)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}

		// users テーブルと同期する。
		user, err := userService.EnsureUser(c.Request.Context(), clerkUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to ensure user",
			})
			c.Abort()
			return
		}

		// 後続のハンドラーから参照できるよう、コンテキストに格納。
		c.Set("user", user)

		// 次のハンドラーに処理を渡す。
		c.Next()
	}
}

package router

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter() *gin.Engine {
	// gin.Default() の代わりに gin.New() を使用し、
	// カスタムのロガーとリカバリーミドルウェアを適用
	r := gin.New()

	// 依存関係の組み立て
	deps := NewDependencies()

	// ミドルウェア設定（ログとリカバリーを含む）
	setupMiddleware(r, deps)

	// ルート設定
	setupRoutes(r, deps)

	return r
}

// setupMiddleware はミドルウェアを設定します。
func setupMiddleware(r *gin.Engine, deps *Dependencies) {
	// リカバリーミドルウェア（パニック時のログ出力）
	r.Use(deps.RecoveryMiddleware)

	// リクエストログミドルウェア（request_id付与、リクエストログ出力）
	r.Use(deps.RequestLoggerMiddleware)

	// CORS設定
	r.Use(cors.New(cors.Config{
		// 許可するオリジン（開発環境と本番環境のフロントエンドURL）
		AllowOrigins: []string{
			"http://localhost:3000",                                // ローカル開発環境
			"http://localhost:8787",                                // ローカル開発環境（Cloudflare Pages プレビュー）
			"https://cinetag-frontend.yuta-develop-ct.workers.dev", // 開発環境（Cloudflare Workers）
		},
		// 許可するHTTPメソッド
		AllowMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		// 許可するリクエストヘッダー（Origin, Content-Type, Authorizationを許可）
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
		// レスポンスでアクセスを許可するヘッダー（Content-Lengthをクライアントに公開）
		ExposeHeaders: []string{"Content-Length"},
		// Cookieなどを含む認証情報のクロスオリジン送信を許可
		AllowCredentials: true,
		// プリフライトリクエスト（OPTIONS）結果のキャッシュ期間（12時間）
		MaxAge: 12 * time.Hour,
	}))
}

// setupRoutes はすべてのルートを設定します。
func setupRoutes(r *gin.Engine, deps *Dependencies) {
	// ヘルスチェック用エンドポイント
	r.GET("/health", healthCheckHandler)

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// API グループ
	api := r.Group("/api/v1")
	{
		// Clerk Webhook
		api.POST("/clerk/webhook", deps.ClerkWebhookHandler.HandleWebhook)

		// 公開ルート（認証不要）
		setupPublicRoutes(api, deps)

		// 認証必須ルート
		setupAuthRoutes(api, deps)
	}
}

// setupPublicRoutes は認証不要の公開ルートを設定します。
func setupPublicRoutes(api *gin.RouterGroup, deps *Dependencies) {
	// タグ（公開）
	api.GET("/tags", deps.TagHandler.ListPublicTags)
	api.GET("/tags/:tagId", deps.OptionalAuthMiddleware, deps.TagHandler.GetTagDetail)
	api.GET("/tags/:tagId/movies", deps.OptionalAuthMiddleware, deps.TagHandler.ListTagMovies)
	api.GET("/tags/:tagId/followers", deps.TagHandler.ListTagFollowers)

	// ユーザー（公開）
	api.GET("/users/:displayId", deps.UserHandler.GetUserByDisplayID)
	api.GET("/users/:displayId/tags", deps.OptionalAuthMiddleware, deps.UserHandler.ListUserTags)
	api.GET("/users/:displayId/following", deps.UserHandler.ListFollowing)
	api.GET("/users/:displayId/followers", deps.UserHandler.ListFollowers)
	api.GET("/users/:displayId/follow-stats", deps.OptionalAuthMiddleware, deps.UserHandler.GetUserFollowStats)

	// 映画（公開）
	api.GET("/movies/search", deps.MovieHandler.SearchMovies)
	api.GET("/movies/:tmdbMovieId", deps.MovieHandler.GetMovieDetail)
	api.GET("/movies/:tmdbMovieId/tags", deps.MovieHandler.GetMovieTags)
}

// setupAuthRoutes は認証必須のルートを設定します。
func setupAuthRoutes(api *gin.RouterGroup, deps *Dependencies) {
	authGroup := api.Group("/")
	authGroup.Use(deps.AuthMiddleware)
	{
		// ユーザー
		setupUserRoutes(authGroup, deps)

		// タグ
		setupTagRoutes(authGroup, deps)

		// タグフォロー
		setupTagFollowRoutes(authGroup, deps)

		// 自分のフォロー中タグ一覧
		authGroup.GET("/me/following-tags", deps.TagHandler.ListFollowingTags)
	}
}

// setupUserRoutes はユーザー関連の認証必須ルートを設定します。
func setupUserRoutes(authGroup *gin.RouterGroup, deps *Dependencies) {
	authGroup.GET("/users/me", deps.UserHandler.GetMe)
	authGroup.PATCH("/users/me", deps.UserHandler.UpdateMe)
	authGroup.POST("/users/:displayId/follow", deps.UserHandler.FollowUser)
	authGroup.DELETE("/users/:displayId/follow", deps.UserHandler.UnfollowUser)
}

// setupTagRoutes はタグ関連の認証必須ルートを設定します。
func setupTagRoutes(authGroup *gin.RouterGroup, deps *Dependencies) {
	authGroup.POST("/tags", deps.TagHandler.CreateTag)
	authGroup.PATCH("/tags/:tagId", deps.TagHandler.UpdateTag)
	authGroup.POST("/tags/:tagId/movies", deps.TagHandler.AddMoviesToTag)
	authGroup.DELETE("/tags/:tagId/movies/:tagMovieId", deps.TagHandler.RemoveMovieFromTag)
}

// setupTagFollowRoutes はタグフォロー関連の認証必須ルートを設定します。
func setupTagFollowRoutes(authGroup *gin.RouterGroup, deps *Dependencies) {
	authGroup.POST("/tags/:tagId/follow", deps.TagHandler.FollowTag)
	authGroup.DELETE("/tags/:tagId/follow", deps.TagHandler.UnfollowTag)
	authGroup.GET("/tags/:tagId/follow-status", deps.TagHandler.GetTagFollowStatus)
}

// healthCheckHandler はヘルスチェック用のハンドラーです。
func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

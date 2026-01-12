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
	r := gin.Default()

	// 依存関係の組み立て
	deps := NewDependencies()

	r.Use(cors.New(cors.Config{
		// 許可するオリジン（開発環境と本番環境のフロントエンドURL）
		AllowOrigins: []string{
			"http://localhost:3000",                                // ローカル開発環境
			"http://localhost:8787",                                // ローカル開発環境（Cloudflare Pages プレビュー）
			"https://cinetag-frontend.yuta-develop-ct.workers.dev", // 本番環境（Cloudflare Workers）
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

	// ヘルスチェック用エンドポイント
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// API グループ
	api := r.Group("/api/v1")
	{

		// Clerk Webhook
		api.POST("/clerk/webhook", deps.ClerkWebhookHandler.HandleWebhook)

		api.GET("/tags", deps.TagHandler.ListPublicTags)
		api.GET("/tags/:tagId", deps.OptionalAuthMiddleware, deps.TagHandler.GetTagDetail)
		api.GET("/tags/:tagId/movies", deps.OptionalAuthMiddleware, deps.TagHandler.ListTagMovies)
		api.GET("/tags/:tagId/followers", deps.TagHandler.ListTagFollowers)

		// ユーザー情報取得（認証不要）
		api.GET("/users/:displayId", deps.UserHandler.GetUserByDisplayID)
		api.GET("/users/:displayId/tags", deps.OptionalAuthMiddleware, deps.UserHandler.ListUserTags)
		api.GET("/users/:displayId/following", deps.UserHandler.ListFollowing)
		api.GET("/users/:displayId/followers", deps.UserHandler.ListFollowers)
		api.GET("/users/:displayId/follow-stats", deps.OptionalAuthMiddleware, deps.UserHandler.GetUserFollowStats)

		// TMDB 検索（認証不要）
		api.GET("/movies/search", deps.MovieHandler.SearchMovies)

		// 認証必須グループ
		authGroup := api.Group("/")
		authGroup.Use(deps.AuthMiddleware)
		{
			// ユーザー
			authGroup.GET("/users/me", deps.UserHandler.GetMe)
			authGroup.POST("/users/:displayId/follow", deps.UserHandler.FollowUser)
			authGroup.DELETE("/users/:displayId/follow", deps.UserHandler.UnfollowUser)

			// タグ
			authGroup.POST("/tags", deps.TagHandler.CreateTag)
			authGroup.PATCH("/tags/:tagId", deps.TagHandler.UpdateTag)
			authGroup.POST("/tags/:tagId/movies", deps.TagHandler.AddMovieToTag)
			authGroup.DELETE("/tags/:tagId/movies/:tagMovieId", deps.TagHandler.RemoveMovieFromTag)

			// タグフォロー
			authGroup.POST("/tags/:tagId/follow", deps.TagHandler.FollowTag)
			authGroup.DELETE("/tags/:tagId/follow", deps.TagHandler.UnfollowTag)
			authGroup.GET("/tags/:tagId/follow-status", deps.TagHandler.GetTagFollowStatus)

			// 自分のフォロー中タグ一覧
			authGroup.GET("/me/following-tags", deps.TagHandler.ListFollowingTags)
		}
	}

	return r
}

package router

import (
	"net/http"
	"os"
	"time"

	"cinetag-backend/src/internal/db"
	"cinetag-backend/src/internal/handler"
	"cinetag-backend/src/internal/middleware"
	"cinetag-backend/src/internal/repository"
	"cinetag-backend/src/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	// 依存関係の組み立て
	database := db.NewDB()
	movieService := service.NewMovieService(database)
	imageBaseURL := os.Getenv("TMDB_IMAGE_BASE_URL")
	tagRepo := repository.NewTagRepository(database)
	tagMovieRepo := repository.NewTagMovieRepository(database)
	tagFollowerRepo := repository.NewTagFollowerRepository(database)
	tagService := service.NewTagService(tagRepo, tagMovieRepo, tagFollowerRepo, movieService, imageBaseURL)
	userRepo := repository.NewUserRepository(database)
	userFollowerRepo := repository.NewUserFollowerRepository(database)
	userService := service.NewUserService(database, userRepo, userFollowerRepo, tagFollowerRepo)
	tagHandler := handler.NewTagHandler(tagService)
	movieHandler := handler.NewMovieHandler(movieService)
	userHandler := handler.NewUserHandler(userService, tagService)
	clerkWebhookHandler := handler.NewClerkWebhookHandler(userService)
	authMiddleware := middleware.NewAuthMiddleware(userService)
	optionalAuthMiddleware := middleware.NewOptionalAuthMiddleware(userService)

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
		api.POST("/clerk/webhook", clerkWebhookHandler.HandleWebhook)

		api.GET("/tags", tagHandler.ListPublicTags)
		api.GET("/tags/:tagId", optionalAuthMiddleware, tagHandler.GetTagDetail)
		api.GET("/tags/:tagId/movies", optionalAuthMiddleware, tagHandler.ListTagMovies)
		api.GET("/tags/:tagId/followers", tagHandler.ListTagFollowers)

		// ユーザー情報取得（認証不要）
		api.GET("/users/:displayId", userHandler.GetUserByDisplayID)
		api.GET("/users/:displayId/tags", optionalAuthMiddleware, userHandler.ListUserTags)
		api.GET("/users/:displayId/following", userHandler.ListFollowing)
		api.GET("/users/:displayId/followers", userHandler.ListFollowers)
		api.GET("/users/:displayId/follow-stats", optionalAuthMiddleware, userHandler.GetUserFollowStats)

		// TMDB 検索（認証不要）
		api.GET("/movies/search", movieHandler.SearchMovies)

		// 認証必須グループ
		authGroup := api.Group("/")
		authGroup.Use(authMiddleware)
		{
			// ユーザー
			authGroup.GET("/users/me", userHandler.GetMe)
			authGroup.POST("/users/:displayId/follow", userHandler.FollowUser)
			authGroup.DELETE("/users/:displayId/follow", userHandler.UnfollowUser)

			// タグ
			authGroup.POST("/tags", tagHandler.CreateTag)
			authGroup.PATCH("/tags/:tagId", tagHandler.UpdateTag)
			authGroup.POST("/tags/:tagId/movies", tagHandler.AddMovieToTag)
			authGroup.DELETE("/tags/:tagId/movies/:tagMovieId", tagHandler.RemoveMovieFromTag)

			// タグフォロー
			authGroup.POST("/tags/:tagId/follow", tagHandler.FollowTag)
			authGroup.DELETE("/tags/:tagId/follow", tagHandler.UnfollowTag)
			authGroup.GET("/tags/:tagId/follow-status", tagHandler.GetTagFollowStatus)

			// 自分のフォロー中タグ一覧
			authGroup.GET("/me/following-tags", tagHandler.ListFollowingTags)
		}
	}

	return r
}

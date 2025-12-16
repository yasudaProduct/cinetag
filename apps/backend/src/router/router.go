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
	tagService := service.NewTagService(tagRepo, tagMovieRepo, movieService, imageBaseURL)
	userRepo := repository.NewUserRepository(database)
	userService := service.NewUserService(userRepo)
	tagHandler := handler.NewTagHandler(tagService)
	clerkWebhookHandler := handler.NewClerkWebhookHandler(userService)
	authMiddleware := middleware.NewAuthMiddleware(userService)

	r.Use(cors.New(cors.Config{
		// 許可するオリジン
		AllowOrigins: []string{"http://localhost:3000"},
		// 許可するHTTPメソッド
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
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
		// 公開タグ一覧（認証不要）
		api.GET("/tags", tagHandler.ListPublicTags)

		// Clerk Webhook
		api.POST("/clerk/webhook", clerkWebhookHandler.HandleWebhook)

		// 認証必須グループ
		authGroup := api.Group("/")
		authGroup.Use(authMiddleware)
		{
			authGroup.POST("/tags", tagHandler.CreateTag)
			authGroup.POST("/tags/:tagId/movies", tagHandler.AddMovieToTag)
		}
	}

	return r
}

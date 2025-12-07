package router

import (
	"net/http"

	"cinetag-backend/src/internal/db"
	"cinetag-backend/src/internal/handler"
	"cinetag-backend/src/internal/middleware"
	"cinetag-backend/src/internal/service"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	// 依存関係の組み立て
	database := db.NewDB()
	tagService := service.NewTagService(database)
	userService := service.NewUserService(database)
	tagHandler := handler.NewTagHandler(tagService)
	clerkWebhookHandler := handler.NewClerkWebhookHandler(userService)
	authMiddleware := middleware.NewAuthMiddleware(userService)

	// ヘルスチェック用エンドポイント
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// すべての API は /api/v1 配下にまとめる
	// api := r.Group("/api/v1")
	// {

	// }

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
		}
	}

	return r
}

package router

import (
	"os"

	"cinetag-backend/src/internal/db"
	"cinetag-backend/src/internal/handler"
	"cinetag-backend/src/internal/middleware"
	"cinetag-backend/src/internal/repository"
	"cinetag-backend/src/internal/service"

	"github.com/gin-gonic/gin"
)

// Dependencies はアプリケーションの依存関係をまとめた構造体です。
type Dependencies struct {
	TagHandler             *handler.TagHandler
	MovieHandler           *handler.MovieHandler
	UserHandler            *handler.UserHandler
	ClerkWebhookHandler    *handler.ClerkWebhookHandler
	AuthMiddleware         gin.HandlerFunc
	OptionalAuthMiddleware gin.HandlerFunc
}

// NewDependencies はアプリケーションの依存関係を組み立てて返します。
func NewDependencies() *Dependencies {
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

	return &Dependencies{
		TagHandler:             tagHandler,
		MovieHandler:           movieHandler,
		UserHandler:            userHandler,
		ClerkWebhookHandler:    clerkWebhookHandler,
		AuthMiddleware:         authMiddleware,
		OptionalAuthMiddleware: optionalAuthMiddleware,
	}
}

package router

import (
	"log/slog"
	"os"

	"cinetag-backend/src/internal/db"
	"cinetag-backend/src/internal/handler"
	"cinetag-backend/src/internal/logger"
	"cinetag-backend/src/internal/middleware"
	"cinetag-backend/src/internal/repository"
	"cinetag-backend/src/internal/service"

	"github.com/gin-gonic/gin"
)

// Dependencies はアプリケーションの依存関係をまとめた構造体です。
type Dependencies struct {
	// Logger は構造化ロガーです。
	Logger *slog.Logger

	// Handlers
	TagHandler          *handler.TagHandler
	MovieHandler        *handler.MovieHandler
	UserHandler         *handler.UserHandler
	ClerkWebhookHandler *handler.ClerkWebhookHandler

	// Middlewares
	RequestLoggerMiddleware gin.HandlerFunc
	RecoveryMiddleware      gin.HandlerFunc
	AuthMiddleware          gin.HandlerFunc
	OptionalAuthMiddleware  gin.HandlerFunc
}

// NewDependencies はアプリケーションの依存関係を組み立てて返します。
func NewDependencies() *Dependencies {
	// Logger の初期化
	log := logger.NewLogger()

	// Database
	database := db.NewDB()

	// Repositories
	tagRepo := repository.NewTagRepository(database)
	tagMovieRepo := repository.NewTagMovieRepository(database)
	tagFollowerRepo := repository.NewTagFollowerRepository(database)
	userRepo := repository.NewUserRepository(database)
	userFollowerRepo := repository.NewUserFollowerRepository(database)

	// Services
	movieService := service.NewMovieService(database)
	imageBaseURL := os.Getenv("TMDB_IMAGE_BASE_URL")
	tagService := service.NewTagService(tagRepo, tagMovieRepo, tagFollowerRepo, movieService, imageBaseURL)
	userService := service.NewUserService(database, userRepo, userFollowerRepo, tagFollowerRepo)

	// Handlers
	tagHandler := handler.NewTagHandler(tagService)
	movieHandler := handler.NewMovieHandler(movieService)
	userHandler := handler.NewUserHandler(userService, tagService)
	clerkWebhookHandler := handler.NewClerkWebhookHandler(userService)

	// Middlewares
	requestLoggerMiddleware := middleware.NewRequestLoggerMiddleware(log)
	recoveryMiddleware := middleware.NewRecoveryMiddleware(log)
	authMiddleware := middleware.NewAuthMiddleware(userService)
	optionalAuthMiddleware := middleware.NewOptionalAuthMiddleware(userService)

	return &Dependencies{
		Logger:                  log,
		TagHandler:              tagHandler,
		MovieHandler:            movieHandler,
		UserHandler:             userHandler,
		ClerkWebhookHandler:     clerkWebhookHandler,
		RequestLoggerMiddleware: requestLoggerMiddleware,
		RecoveryMiddleware:      recoveryMiddleware,
		AuthMiddleware:          authMiddleware,
		OptionalAuthMiddleware:  optionalAuthMiddleware,
	}
}

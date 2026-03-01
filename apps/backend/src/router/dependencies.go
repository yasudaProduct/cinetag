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
	NotificationHandler *handler.NotificationHandler
	ClerkWebhookHandler *handler.ClerkWebhookHandler

	// Middlewares
	MaintenanceMiddleware   gin.HandlerFunc
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
	userRepo := repository.NewUserRepository(log, database)
	userFollowerRepo := repository.NewUserFollowerRepository(database)

	// Services
	movieService := service.NewMovieService(log, database)
	notifRepo := repository.NewNotificationRepository(database)
	notificationService := service.NewNotificationService(log, notifRepo, tagRepo, tagFollowerRepo, userFollowerRepo)
	imageBaseURL := os.Getenv("TMDB_IMAGE_BASE_URL")
	tagService := service.NewTagService(log, tagRepo, tagMovieRepo, tagFollowerRepo, movieService, notificationService, imageBaseURL)
	userService := service.NewUserService(log, database, userRepo, userFollowerRepo, tagFollowerRepo, notificationService)

	// Handlers
	tagHandler := handler.NewTagHandler(log, tagService)
	movieHandler := handler.NewMovieHandler(log, movieService)
	userHandler := handler.NewUserHandler(log, userService, tagService)
	notificationHandler := handler.NewNotificationHandler(log, notificationService)
	clerkWebhookHandler := handler.NewClerkWebhookHandler(log, userService)

	// Middlewares
	maintenanceMiddleware := middleware.NewMaintenanceMiddleware(log)
	requestLoggerMiddleware := middleware.NewRequestLoggerMiddleware(log)
	recoveryMiddleware := middleware.NewRecoveryMiddleware(log)
	authMiddleware := middleware.NewAuthMiddleware(log, userService)
	optionalAuthMiddleware := middleware.NewOptionalAuthMiddleware(log, userService)

	return &Dependencies{
		Logger:                  log,
		TagHandler:              tagHandler,
		MovieHandler:            movieHandler,
		UserHandler:             userHandler,
		NotificationHandler:     notificationHandler,
		ClerkWebhookHandler:     clerkWebhookHandler,
		MaintenanceMiddleware:   maintenanceMiddleware,
		RequestLoggerMiddleware: requestLoggerMiddleware,
		RecoveryMiddleware:      recoveryMiddleware,
		AuthMiddleware:          authMiddleware,
		OptionalAuthMiddleware:  optionalAuthMiddleware,
	}
}

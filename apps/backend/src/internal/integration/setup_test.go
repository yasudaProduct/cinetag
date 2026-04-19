//go:build integration

package integration

import (
	"context"
	"database/sql"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"cinetag-backend/src/internal/handler"
	"cinetag-backend/src/internal/migration"
	"cinetag-backend/src/internal/model"
	"cinetag-backend/src/internal/repository"
	"cinetag-backend/src/internal/service"
	"cinetag-backend/src/internal/testutil"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// testEnv は API 結合テストで共有する環境を保持します。
type testEnv struct {
	db        *gorm.DB
	router    *gin.Engine
	testUsers map[string]*model.User
}

// setupTestEnv はテスト全体で1回だけ呼び出し、DB・ルーター・テストユーザーを初期化します。
func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL が未設定のため integration テストをスキップします")
	}

	db := openDB(t, dsn)
	applyMigrations(t, dsn)
	truncateAll(t, db)

	testUsers := make(map[string]*model.User)
	router := buildRouter(t, db, testUsers)

	return &testEnv{
		db:        db,
		router:    router,
		testUsers: testUsers,
	}
}

func openDB(t *testing.T, dsn string) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("DB 接続に失敗: %v", err)
	}
	return db
}

// applyMigrations は goose でマイグレーションを最新まで適用します。
func applyMigrations(t *testing.T, dsn string) {
	t.Helper()

	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("goose 用の DB 接続に失敗: %v", err)
	}
	defer sqlDB.Close()

	migrationsFS, err := fs.Sub(migration.Migrations, "migrations")
	if err != nil {
		t.Fatalf("migration FS の取得に失敗: %v", err)
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		sqlDB,
		migrationsFS,
	)
	if err != nil {
		t.Fatalf("goose provider の生成に失敗: %v", err)
	}

	if _, err := provider.Up(context.Background()); err != nil {
		t.Fatalf("goose up に失敗: %v", err)
	}
}

// truncateAll は全テーブルを TRUNCATE して各テストの独立性を確保します。
// 外部キーの依存関係を考慮した順序になっています。
func truncateAll(t *testing.T, db *gorm.DB) {
	t.Helper()
	tables := []string{
		"notifications",
		"tag_likes",
		"tag_followers",
		"tag_movies",
		"user_followers",
		"tags",
		"movie_cache",
		"users",
	}
	for _, tbl := range tables {
		if err := db.Exec("TRUNCATE TABLE " + tbl + " CASCADE").Error; err != nil {
			t.Fatalf("TRUNCATE %s に失敗: %v", tbl, err)
		}
	}
}

// buildRouter はテスト用に本番と同じルート構造を持つ gin.Engine を組み立てます。
// DB は引数で渡し、認証はバイパスミドルウェアを使います。
func buildRouter(t *testing.T, db *gorm.DB, testUsers map[string]*model.User) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Repositories
	tagRepo := repository.NewTagRepository(db)
	tagMovieRepo := repository.NewTagMovieRepository(db)
	tagFollowerRepo := repository.NewTagFollowerRepository(db)
	tagLikeRepo := repository.NewTagLikeRepository(db)
	userRepo := repository.NewUserRepository(log, db)
	userFollowerRepo := repository.NewUserFollowerRepository(db)
	notifRepo := repository.NewNotificationRepository(db)

	// Services
	movieService := service.NewMovieService(log, db)
	notificationService := service.NewNotificationService(log, notifRepo, tagRepo, tagFollowerRepo, userFollowerRepo)
	tagService := service.NewTagService(log, tagRepo, tagMovieRepo, tagFollowerRepo, tagLikeRepo, movieService, notificationService, "")
	userService := service.NewUserService(log, db, userRepo, userFollowerRepo, tagFollowerRepo, notificationService)

	// Handlers
	tagHandler := handler.NewTagHandler(log, tagService)
	movieHandler := handler.NewMovieHandler(log, movieService)
	userHandler := handler.NewUserHandler(log, userService, tagService)
	notificationHandler := handler.NewNotificationHandler(log, notificationService)
	clerkWebhookHandler := handler.NewClerkWebhookHandler(log, userService)

	// Auth bypass middlewares
	authMW := testutil.TestAuthMiddleware(testUsers)
	optionalAuthMW := testutil.TestOptionalAuthMiddleware(testUsers)

	r := gin.New()

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		api.POST("/clerk/webhook", clerkWebhookHandler.HandleWebhook)

		// 公開ルート（認証不要）
		api.GET("/tags", tagHandler.ListPublicTags)
		api.GET("/tags/:tagId", optionalAuthMW, tagHandler.GetTagDetail)
		api.GET("/tags/:tagId/movies", optionalAuthMW, tagHandler.ListTagMovies)
		api.GET("/tags/:tagId/followers", tagHandler.ListTagFollowers)

		api.GET("/users/:displayId", userHandler.GetUserByDisplayID)
		api.GET("/users/:displayId/tags", optionalAuthMW, userHandler.ListUserTags)
		api.GET("/users/:displayId/following", userHandler.ListFollowing)
		api.GET("/users/:displayId/followers", userHandler.ListFollowers)
		api.GET("/users/:displayId/follow-stats", optionalAuthMW, userHandler.GetUserFollowStats)

		api.GET("/movies/search", movieHandler.SearchMovies)
		api.GET("/movies/:tmdbMovieId", movieHandler.GetMovieDetail)
		api.GET("/movies/:tmdbMovieId/tags", movieHandler.GetMovieTags)

		// 認証必須ルート
		auth := api.Group("/")
		auth.Use(authMW)
		{
			auth.GET("/users/me", userHandler.GetMe)
			auth.PATCH("/users/me", userHandler.UpdateMe)
			auth.POST("/users/:displayId/follow", userHandler.FollowUser)
			auth.DELETE("/users/:displayId/follow", userHandler.UnfollowUser)

			auth.POST("/tags", tagHandler.CreateTag)
			auth.PATCH("/tags/:tagId", tagHandler.UpdateTag)
			auth.DELETE("/tags/:tagId", tagHandler.DeleteTag)
			auth.POST("/tags/:tagId/movies", tagHandler.AddMoviesToTag)
			auth.DELETE("/tags/:tagId/movies/:tagMovieId", tagHandler.RemoveMovieFromTag)

			auth.POST("/tags/:tagId/follow", tagHandler.FollowTag)
			auth.DELETE("/tags/:tagId/follow", tagHandler.UnfollowTag)
			auth.GET("/tags/:tagId/follow-status", tagHandler.GetTagFollowStatus)

			auth.POST("/tags/:tagId/like", tagHandler.LikeTag)
			auth.DELETE("/tags/:tagId/like", tagHandler.UnlikeTag)
			auth.GET("/tags/:tagId/like-status", tagHandler.GetTagLikeStatus)

			auth.GET("/notifications", notificationHandler.ListNotifications)
			auth.GET("/notifications/unread-count", notificationHandler.GetUnreadCount)
			auth.PATCH("/notifications/:notificationId/read", notificationHandler.MarkAsRead)
			auth.PATCH("/notifications/read-all", notificationHandler.MarkAllAsRead)

			auth.GET("/me/following-tags", tagHandler.ListFollowingTags)
			auth.GET("/me/liked-tags", tagHandler.ListLikedTags)
		}
	}

	return r
}

// createUser はテスト用ユーザーを DB に直接作成し、testUsers マップにも登録します。
func (e *testEnv) createUser(t *testing.T, clerkID, displayID, displayName string) *model.User {
	t.Helper()
	u := &model.User{
		ClerkUserID: clerkID,
		DisplayID:   displayID,
		DisplayName: displayName,
		Email:       displayName + "@example.com",
	}
	if err := e.db.Create(u).Error; err != nil {
		t.Fatalf("テストユーザー作成に失敗: %v", err)
	}
	e.testUsers[u.ID] = u
	return u
}

// request は testutil.PerformRequest のラッパーです。
func (e *testEnv) request(method, path string, body []byte, headers map[string]string) *testutil.HTTPResponse {
	rw := testutil.PerformRequest(e.router, method, path, body, headers)
	return &testutil.HTTPResponse{Recorder: rw}
}

// authHeaders は認証バイパス用のヘッダーを返します。
func authHeaders(userID string) map[string]string {
	return map[string]string{
		"X-Test-User-ID": userID,
		"Content-Type":   "application/json",
	}
}

// jsonHeaders は Content-Type のみのヘッダーを返します。
func jsonHeaders() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

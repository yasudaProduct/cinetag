package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"

	"cinetag-backend/src/internal/migration"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// サブコマンドの解析
	// 使い方: go run ./src/cmd/migrate [up|down|status|reset]
	command := "up"
	if len(os.Args) > 1 {
		command = strings.ToLower(os.Args[1])
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// embed.FS は "migrations/*.sql" のパスで埋め込まれるため、
	// fs.Sub でサブディレクトリをルートにする
	migrationsFS, err := fs.Sub(migration.Migrations, "migrations")
	if err != nil {
		log.Fatalf("failed to get migrations sub-filesystem: %v", err)
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		migrationsFS,
		goose.WithVerbose(true),
	)
	if err != nil {
		log.Fatalf("failed to create goose provider: %v", err)
	}

	ctx := context.Background()

	switch command {
	case "up":
		runUp(ctx, provider)
	case "down":
		runDown(ctx, provider)
	case "status":
		runStatus(ctx, provider)
	case "reset":
		runReset(ctx, provider)
	default:
		log.Fatalf("unknown command: %s (use: up, down, status, reset)", command)
	}

	log.Printf("migration '%s' completed successfully", command)
}

// runUp は未適用のマイグレーションを全て適用します。
func runUp(ctx context.Context, provider *goose.Provider) {
	results, err := provider.Up(ctx)
	if err != nil {
		log.Fatalf("migration up failed: %v", err)
	}
	for _, r := range results {
		log.Printf("applied: %s (%s)", r.Source.Path, r.Duration)
	}
	if len(results) == 0 {
		log.Println("no new migrations to apply")
	}
}

// runDown は最新のマイグレーション1つをロールバックします。
func runDown(ctx context.Context, provider *goose.Provider) {
	result, err := provider.Down(ctx)
	if err != nil {
		log.Fatalf("migration down failed: %v", err)
	}
	if result == nil {
		log.Println("no migrations to roll back")
		return
	}
	log.Printf("rolled back: %s (%s)", result.Source.Path, result.Duration)
}

// runStatus は全マイグレーションの適用状況を表示します。
func runStatus(ctx context.Context, provider *goose.Provider) {
	results, err := provider.Status(ctx)
	if err != nil {
		log.Fatalf("migration status failed: %v", err)
	}
	fmt.Printf("%-10s %-50s %s\n", "VERSION", "NAME", "STATUS")
	fmt.Println(strings.Repeat("-", 80))
	for _, r := range results {
		state := "Pending"
		if r.State == goose.StateApplied {
			state = fmt.Sprintf("Applied at %s", r.AppliedAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Printf("%-10d %-50s %s\n", r.Source.Version, r.Source.Path, state)
	}
}

// runReset は全マイグレーションをロールバックしてから再適用します。
// ENV=develop の場合のみ実行可能です。
func runReset(ctx context.Context, provider *goose.Provider) {
	env := strings.TrimSpace(strings.ToLower(os.Getenv("ENV")))
	if env != "develop" {
		log.Fatal("reset is only allowed in develop environment (ENV=develop)")
	}

	log.Println("ENV=develop: resetting all migrations...")

	downResults, err := provider.DownTo(ctx, 0)
	if err != nil {
		log.Fatalf("migration down-to-0 failed: %v", err)
	}
	for _, r := range downResults {
		log.Printf("rolled back: %s (%s)", r.Source.Path, r.Duration)
	}

	upResults, err := provider.Up(ctx)
	if err != nil {
		log.Fatalf("migration up failed: %v", err)
	}
	for _, r := range upResults {
		log.Printf("applied: %s (%s)", r.Source.Path, r.Duration)
	}
}

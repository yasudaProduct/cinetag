package main

import (
	"log"
	"os"
	"strings"

	"cinetag-backend/src/internal/db"
	"cinetag-backend/src/internal/model"

	"gorm.io/gorm"
)

// resetSchemaIfEnabled は、開発環境でのみ「全テーブル削除（public スキーマ再作成）」を実行します。
//
// 本番での誤実行を防ぐため、ENV=develop の場合のみ実行します。
func resetSchemaIfEnabled(database *gorm.DB) {
	env := strings.TrimSpace(strings.ToLower(os.Getenv("ENV")))
	if env != "develop" {
		return
	}

	log.Println("ENV=develop: resetting schema (DROP SCHEMA public CASCADE; CREATE SCHEMA public;)")
	if err := database.Exec(`DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;`).Error; err != nil {
		log.Fatalf("failed to reset schema: %v", err)
	}
}

// このコマンドはデータベースマイグレーション専用のエントリーポイントです。
// アプリケーション本体とは別に実行し、スキーマ更新のみを行います。
func main() {
	database := db.NewDB()

	resetSchemaIfEnabled(database)

	if err := database.AutoMigrate(
		&model.User{},
		&model.Tag{},
		&model.TagMovie{},
		&model.TagFollower{},
		&model.MovieCache{},
	); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	log.Println("migration completed successfully")
}

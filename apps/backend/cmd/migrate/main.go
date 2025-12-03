package main

import (
	"log"

	"cinetag-backend/internal/db"
	"cinetag-backend/internal/model"
)

// このコマンドはデータベースマイグレーション専用のエントリーポイントです。
// アプリケーション本体とは別に実行し、スキーマ更新のみを行います。
func main() {
	database := db.NewDB()

	if err := database.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.CategoryMovie{},
		&model.CategoryFollower{},
		&model.MovieCache{},
	); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	log.Println("migration completed successfully")
}



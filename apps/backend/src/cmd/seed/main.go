package main

import (
	"log"

	"cinetag-backend/src/internal/db"
	"cinetag-backend/src/internal/seed"
)

// このコマンドは開発用シードデータ投入の専用エントリーポイントです。
// ENV=develop の場合のみデータを投入します。
//
// 使い方: ENV=develop go run ./src/cmd/seed
func main() {
	database := db.NewDB()

	if err := seed.SeedDevelop(database); err != nil {
		log.Fatalf("failed to seed database: %v", err)
	}

	log.Println("seed completed successfully")
}

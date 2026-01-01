package db

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		// panic(".envファイルの情報が取得できません")
		log.Println(".envファイルの情報が取得できません")
	}
}

// NewDB はアプリケーションで利用する *gorm.DB を初期化して返します。
// この関数は「接続の確立」のみに責務を持ち、マイグレーションは別コマンドで実行します。
func NewDB() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	return db
}

package main

import (
	"log"
	"os"

	docs "cinetag-backend/src/cmd/docs"
	appRouter "cinetag-backend/src/router"
)

func main() {
	// ポート番号の取得
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Swagger メタ情報の設定
	docs.SwaggerInfo.Host = "localhost:" + port
	docs.SwaggerInfo.BasePath = "/api/v1"

	// ルーターの初期化
	router := appRouter.NewRouter()

	// サーバーの起動
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

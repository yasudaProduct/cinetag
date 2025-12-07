package main

import (
	"log"
	"os"

	appRouter "cinetag-backend/src/router"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := appRouter.NewRouter()

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}



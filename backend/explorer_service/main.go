package main

import (
	"log"
	"os"

	"explorer_service/config"
	"explorer_service/delivery/http"
	"explorer_service/repository"
	"explorer_service/usecase"
	"explorer_service/worker"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: file .env tidak ditemukan")
	}

	// 1. Inisialisasi Database
	config.ConnectDatabase()
	db := config.DB

	// 2. Inisialisasi Repositories
	blockRepo := repository.NewBlockRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	settingRepo := repository.NewSettingRepository(db)

	// 3. Inisialisasi Usecase
	explorerUsecase := usecase.NewExplorerUsecase(blockRepo, txRepo)

	// 4. Start Background Worker
	syncWorker := worker.NewSyncWorker(blockRepo, txRepo, settingRepo)
	syncWorker.Start()

	// 5. Inisialisasi HTTP Server
	r := gin.Default()

	// CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	apiV1 := r.Group("/api/v1")

	// 6. Inisialisasi Handlers
	http.NewExplorerHandler(apiV1, explorerUsecase)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Explorer Service berjalan di port :%s\n", port)
	r.Run(":" + port)
}

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

	// 1. Inisialisasi Database (PostgreSQL untuk Settings)
	config.ConnectDatabase()
	db := config.DB

	// 2. Inisialisasi Blockchain Repository (Ganache)
	ganacheURL := os.Getenv("GANACHE_URL")
	if ganacheURL == "" {
		ganacheURL = "http://ganache:8545"
	}
	systemPrivateKey := os.Getenv("SYSTEM_PRIVATE_KEY")
	if systemPrivateKey == "" {
		systemPrivateKey = "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"
	}
	blockchainRepo, err := repository.NewBlockchainRepository(ganacheURL, systemPrivateKey, db)
	if err != nil {
		log.Fatalf("Gagal terhubung ke Ganache: %v", err)
	}

	settingRepo := repository.NewSettingRepository(db)

	// 3. Inisialisasi Usecase
	explorerUsecase := usecase.NewExplorerUsecase(blockchainRepo)

	// 4. Inisialisasi Message Broker
	amqpURL := os.Getenv("RABBITMQ_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@rabbitmq:5672/"
	}
	broker, err := worker.NewBroker(amqpURL)
	if err != nil {
		log.Fatalf("Gagal terhubung ke RabbitMQ: %v", err)
	}
	defer broker.Close()

	// 5. Start Background Worker
	syncWorker := worker.NewSyncWorker(blockchainRepo, settingRepo, broker)
	syncWorker.Start()

	// 6. Inisialisasi HTTP Server
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

	// 7. Inisialisasi Handlers
	http.NewExplorerHandler(apiV1, explorerUsecase)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Explorer Service berjalan di port :%s\n", port)
	r.Run(":" + port)
}


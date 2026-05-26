package main

import (
	"log"
	"os"

	"backend/config"
	"backend/delivery/http"
	"backend/repository"
	"backend/usecase"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: file .env tidak ditemukan")
	}

	// 1. Inisialisasi Database (Driver & ORM)
	config.ConnectDatabase()
	db := config.DB

	// 2. Inisialisasi Repositories (Layer 2)
	userRepo := repository.NewUserRepository(db)
	settingRepo := repository.NewSettingRepository(db)

	// 3. Inisialisasi Usecases (Layer 3)
	authUsecase := usecase.NewAuthUsecase(userRepo)
	settingUsecase := usecase.NewSettingUsecase(settingRepo)

	// 4. Inisialisasi HTTP Server (Layer 4)
	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	apiV1 := r.Group("/api/v1")

	// 5. Inisialisasi Handlers (Delivery)
	http.NewAuthHandler(apiV1, authUsecase)
	http.NewSettingHandler(apiV1, settingUsecase)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server berjalan di port :%s\n", port)
	r.Run(":" + port)
}

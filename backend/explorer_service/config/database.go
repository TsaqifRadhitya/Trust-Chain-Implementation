package config

import (
	"fmt"
	"log"
	"os"

	"explorer_service/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	if host == "" {
		host = "localhost"
		port = "5432"
		user = "postgres"
		password = "postgres"
		dbname = "auth_service"
		sslmode = "disable"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal connect ke database: %v", err)
	}

	// Auto Migrate
	err = db.AutoMigrate(&domain.Block{}, &domain.Transaction{})
	if err != nil {
		log.Fatalf("Gagal migrasi database: %v", err)
	}

	DB = db
	log.Println("Berhasil connect dan migrasi database!")
}

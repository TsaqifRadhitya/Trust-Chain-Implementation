package config

import (
	"fmt"
	"log"
	"os"

	"backend/domain"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "auth_service"
	}
	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	// 1. Hubungkan ke database default 'postgres' terlebih dahulu untuk memeriksa/membuat database target
	defaultDsn := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s sslmode=%s", host, user, password, port, sslmode)
	defaultDb, err := gorm.Open(postgres.Open(defaultDsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal terhubung ke PostgreSQL default server: %v", err)
	}

	// Periksa apakah database target sudah ada
	var exists bool
	err = defaultDb.Raw("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = ?)", dbname).Scan(&exists).Error
	if err != nil {
		log.Fatalf("Gagal memeriksa keberadaan database target: %v", err)
	}

	if !exists {
		log.Printf("Database '%s' tidak ditemukan. Membuat database baru...\n", dbname)
		// Jalankan CREATE DATABASE (tidak boleh dalam transaksi)
		err = defaultDb.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname)).Error
		if err != nil {
			log.Fatalf("Gagal membuat database target '%s': %v", dbname, err)
		}
		log.Printf("Database '%s' berhasil dibuat!\n", dbname)
	}

	// Tutup koneksi database default
	sqlDb, err := defaultDb.DB()
	if err == nil {
		sqlDb.Close()
	}

	// 2. Hubungkan ke database target
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", host, user, password, dbname, port, sslmode)
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal terhubung ke database target '%s': %v", dbname, err)
	}

	log.Printf("Koneksi database PostgreSQL '%s' berhasil!\n", dbname)

	// Jalankan AutoMigrate
	err = DB.AutoMigrate(&domain.User{}, &domain.Configuration{})
	if err != nil {
		log.Fatalf("Gagal menjalankan AutoMigrate: %v", err)
	}
	log.Println("Migrasi skema database berhasil!")

	// Jalankan Seeder
	SeedDatabase()
}

func SeedDatabase() {
	log.Println("Memulai verifikasi seeding database...")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Gagal melakukan hashing password seeder: %v\n", err)
		return
	}

	adminUser := domain.User{
		Email:    "admin@trustchain.com",
		Name:     "Admin TrustChain",
		Password: string(hashedPassword),
		Role:     "admin",
	}

	superAdminUser := domain.User{
		Email:    "superadmin@trustchain.com",
		Name:     "Super Admin TrustChain",
		Password: string(hashedPassword),
		Role:     "superadmin",
	}

	err = DB.Where(domain.User{Email: adminUser.Email}).FirstOrCreate(&adminUser).Error
	if err != nil {
		log.Printf("Gagal seeding admin user: %v\n", err)
	} else {
		log.Printf("Seeding admin user terverifikasi! (Email: %s)\n", adminUser.Email)
	}

	err = DB.Where(domain.User{Email: superAdminUser.Email}).FirstOrCreate(&superAdminUser).Error
	if err != nil {
		log.Printf("Gagal seeding super admin user: %v\n", err)
	} else {
		log.Printf("Seeding super admin user terverifikasi! (Email: %s)\n", superAdminUser.Email)
	}

	// Seed default settings untuk user admin ini
	var configCount int64
	DB.Model(&domain.Configuration{}).Where("user_id = ?", adminUser.ID).Count(&configCount)
	if configCount == 0 && adminUser.ID != 0 {
		defaultConfig := domain.Configuration{
			UserID:            adminUser.ID,
			ErpType:           "SAP S/4HANA",
			Endpoint:          "https://erp.internal.company.com/api/v2/transactions",
			ApiKey:            "sk-trustchain-default-placeholder",
			VolumeSensitivity: 85,
			GeoThreshold:      50,
			VelocityLimit:     70,
		}

		err = DB.Create(&defaultConfig).Error
		if err != nil {
			log.Printf("Gagal seeding default configuration: %v\n", err)
		} else {
			log.Println("Seeding default configuration berhasil!")
		}
	}
}

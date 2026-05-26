package main

import (
	"log"

	"backend/config"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env jika ada
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: file .env tidak ditemukan")
	}

	// ConnectDatabase secara otomatis akan memanggil config.SeedDatabase()
	// dan melakukan auto migrate karena logic-nya ada di dalam fungsi tersebut.
	log.Println("Menjalankan inisialisasi database dan seeder...")
	config.ConnectDatabase()
	log.Println("Proses seeder selesai!")
}

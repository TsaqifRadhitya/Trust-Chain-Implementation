package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type TransactionInput struct {
	VendorName            string  `json:"vendor_name"`
	AmountIDR             float64 `json:"amount_idr"`
	HourOfDay             int     `json:"hour_of_day"`
	DayOfWeek             int     `json:"day_of_week"`
	IsWeekend             int     `json:"is_weekend"`
	VendorAgeDays         int     `json:"vendor_age_days"`
	VendorTxCount30d      int     `json:"vendor_tx_count_30d"`
	AmountVsVendorAvg     float64 `json:"amount_vs_vendor_avg"`
	GeographicDeviation   float64 `json:"geographic_deviation"`
	TxVelocity1h          int     `json:"tx_velocity_1h"`
	TxVelocity24h         int     `json:"tx_velocity_24h"`
	IsRoundNumber         int     `json:"is_round_number"`
	DaysSinceLastTxVendor int     `json:"days_since_last_tx_vendor"`
	IpCountryMatch        int     `json:"ip_country_match"`
	DuplicateScore        float64 `json:"duplicate_score"`
	VendorCategory        string  `json:"vendor_category"`
	Department            string  `json:"department"`
	TransactionType       string  `json:"transaction_type"`
	PaymentMethod         string  `json:"payment_method"`
	ApprovalLevel         string  `json:"approval_level"`
}

// --- Helper functions untuk mempermudah randomisasi ---
func randomInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func randomFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func predictHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vendors := []string{"Neo Supply International", "Global Tech Corp", "PT Sinar Logistik", "Sunrise Raw Materials", "Acme Corp"}
	categories := []string{"Trading", "Logistics", "Manufacturing", "Energy", "Chemicals", "Construction", "Electronics", "Raw Materials", "Engineering"}
	
	var output TransactionInput
	output.VendorName = vendors[rand.Intn(len(vendors))]
	output.VendorCategory = categories[rand.Intn(len(categories))]

	// ── MENENTUKAN PELUANG FRAUD (Misal: 10% Fraud, 90% Normal) ──
	isFraud := rand.Float64() < 0.10

	if isFraud {
		// ==========================================
		// POLA FRAUD (10% Kasus)
		// ==========================================
		output.AmountIDR = randomFloat(1_000_000_000, 10_000_000_000) // 1 Milyar - 10 Milyar
		output.HourOfDay = randomInt(0, 4)                            // Dini hari (jam 0-4)
		output.DayOfWeek = randomInt(0, 6)
		output.IsWeekend = 1                                          // Biasanya dieksekusi saat libur/weekend
		output.VendorAgeDays = randomInt(1, 60)                       // Vendor baru (1-60 hari)
		output.VendorTxCount30d = randomInt(10, 50)                   // Sangat sering transaksi akhir-akhir ini
		output.AmountVsVendorAvg = randomFloat(3.0, 10.0)             // 3x sampai 10x dari rata-rata vendor
		output.GeographicDeviation = randomFloat(0.6, 1.0)            // Lokasi aneh/jauh
		output.TxVelocity1h = randomInt(5, 20)                        // Velocity tinggi (Smurfing)
		output.TxVelocity24h = randomInt(15, 50)
		output.IsRoundNumber = 1                                      // Angka bulat (misal pas 5 Milyar)
		output.DaysSinceLastTxVendor = randomInt(0, 1)                // Tidak ada jeda istirahat
		output.IpCountryMatch = 0                                     // IP luar negeri / VPN
		output.DuplicateScore = randomFloat(0.4, 0.9)                 // Mirip dengan transaksi lain

		// Konteks Organisasi berisiko tinggi
		output.Department = "Finance"
		output.TransactionType = "Advance Payment"
		output.PaymentMethod = "SWIFT"
		output.ApprovalLevel = "L4"

		// Cetak log di server Golang untuk memudahkan monitoring
		fmt.Println("⚠️  MENGHASILKAN DATA DUMMY FRAUD!")

	} else {
		// ==========================================
		// POLA NORMAL (90% Kasus)
		// ==========================================
		output.AmountIDR = randomFloat(5_000_000, 100_000_000)        // 5 Juta - 100 Juta
		output.HourOfDay = randomInt(8, 17)                           // Jam kerja kantor (8 pagi - 5 sore)
		output.DayOfWeek = randomInt(1, 5)                            // Senin - Jumat
		output.IsWeekend = 0
		output.VendorAgeDays = randomInt(365, 2000)                   // Vendor lama terpercaya (> 1 tahun)
		output.VendorTxCount30d = randomInt(1, 10)                    // Wajar (1-10 transaksi sebulan)
		output.AmountVsVendorAvg = randomFloat(0.8, 1.2)              // Sangat dekat dengan rata-rata (0.8x - 1.2x)
		output.GeographicDeviation = randomFloat(0.0, 0.1)            // Lokasi wajar
		output.TxVelocity1h = randomInt(0, 1)                         // Lambat/Wajar
		output.TxVelocity24h = randomInt(1, 3)
		output.IsRoundNumber = 0                                      // Angka acak/keriting (karena pajak dll)
		output.DaysSinceLastTxVendor = randomInt(5, 30)               // Jeda waktu panjang
		output.IpCountryMatch = 1                                     // IP Indonesia / Cocok
		output.DuplicateScore = randomFloat(0.0, 0.05)                // Skor duplikasi sangat rendah

		// Konteks Organisasi berisiko rendah/standar
		normalDepts := []string{"Procurement", "Operations", "HR", "Logistics"}
		normalTx := []string{"Invoice Payment", "Reimbursement", "Purchase Order"}
		normalPay := []string{"Bank Transfer", "RTGS", "Virtual Account"}
		normalAppv := []string{"L1", "L2"}

		output.Department = normalDepts[rand.Intn(len(normalDepts))]
		output.TransactionType = normalTx[rand.Intn(len(normalTx))]
		output.PaymentMethod = normalPay[rand.Intn(len(normalPay))]
		output.ApprovalLevel = normalAppv[rand.Intn(len(normalAppv))]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "online",
		"message": "Dummy Golang API (ERP) is running with Smart Data Generation!",
	})
}

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/predict", predictHandler)
	http.HandleFunc("/", healthHandler)

	port := "8080"
	fmt.Printf("Dummy API running on http://localhost:%s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
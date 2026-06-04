package worker

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"explorer_service/domain"
	"explorer_service/repository"
)

type SyncWorker struct {
	blockchainRepo repository.BlockchainRepository
	settingRepo    repository.SettingRepository
	broker         *Broker
}

func NewSyncWorker(
	b repository.BlockchainRepository,
	s repository.SettingRepository,
	broker *Broker,
) *SyncWorker {
	return &SyncWorker{
		blockchainRepo: b,
		settingRepo:    s,
		broker:         broker,
	}
}

// Start menjalankan tiga goroutine:
//  1. Scheduler: setiap 30 menit publish SyncJob per konfigurasi ke queue.
//  2. Consumer SyncJobs: memproses SyncJob dari queue satu per satu.
//  3. Consumer PredictResponses: menerima hasil prediksi model_service secara asinkron.
func (w *SyncWorker) Start() {
	log.Println("[Sync Worker] Memulai Background Sync Worker dengan RabbitMQ...")

	// Goroutine consumer sync_jobs — berjalan selamanya
	go w.startConsumer()

	// Goroutine consumer predict_responses — berjalan selamanya
	go w.startPredictResponseConsumer()

	// Goroutine scheduler — trigger pertama langsung, lalu setiap 30 menit
	go func() {
		w.publishAllJobs()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			w.publishAllJobs()
		}
	}()
}

// publishAllJobs membaca semua konfigurasi dan mengirim SyncJob ke queue.
func (w *SyncWorker) publishAllJobs() {
	log.Println("[Sync Worker] Mengambil konfigurasi ERP dari database...")

	confs, err := w.settingRepo.GetAllConfigurations()
	if err != nil {
		log.Printf("[Sync Worker] Error mengambil konfigurasi: %v\n", err)
		return
	}

	if len(confs) == 0 {
		log.Println("[Sync Worker] Tidak ada konfigurasi di database. Sync dilewati.")
		return
	}

	for _, conf := range confs {
		endpoint := strings.TrimSpace(conf.Endpoint)
		if endpoint == "" {
			continue
		}

		job := SyncJob{
			UserID:   conf.UserID,
			ConfigID: conf.ID,
			Endpoint: endpoint,
			ApiKey:   conf.ApiKey,
		}

		if err := w.broker.Publish(job); err != nil {
			log.Printf("[Sync Worker] Gagal publish SyncJob untuk UserID %d: %v\n", conf.UserID, err)
			continue
		}
		log.Printf("[Sync Worker] SyncJob dipublish untuk UserID %d (endpoint: %s)\n", conf.UserID, endpoint)
	}
}

// startConsumer memulai loop konsumsi pesan dari RabbitMQ.
func (w *SyncWorker) startConsumer() {
	msgs, err := w.broker.Consume()
	if err != nil {
		log.Fatalf("[Sync Worker] Gagal memulai consumer RabbitMQ: %v\n", err)
	}

	log.Println("[Sync Worker] Consumer RabbitMQ aktif, menunggu SyncJob...")

	for msg := range msgs {
		w.handleMessage(msg)
	}

	log.Println("[Sync Worker] Consumer RabbitMQ berhenti (channel ditutup).")
}

// handleMessage mendekode pesan dan memproses satu SyncJob.
func (w *SyncWorker) handleMessage(msg amqp.Delivery) {
	var job SyncJob
	if err := json.Unmarshal(msg.Body, &job); err != nil {
		log.Printf("[Sync Worker] Pesan tidak valid, dibuang: %v\n", err)
		msg.Nack(false, false) // buang, jangan requeue
		return
	}

	log.Printf("[Sync Worker] Memproses SyncJob — UserID: %d, Endpoint: %s\n", job.UserID, job.Endpoint)

	// Ambil konfigurasi lengkap dari DB (termasuk sensitivitas parameter model)
	conf, err := w.settingRepo.GetConfigurationByUserID(job.UserID)
	if err != nil {
		log.Printf("[Sync Worker] Konfigurasi UserID %d tidak ditemukan: %v\n", job.UserID, err)
		msg.Nack(false, false)
		return
	}

	if err := w.processSingleConfig(conf, job.Endpoint, job.ApiKey); err != nil {
		log.Printf("[Sync Worker] Gagal memproses SyncJob UserID %d: %v\n", job.UserID, err)
		msg.Nack(false, false) // gagal, jangan requeue (hindari loop)
		return
	}

	msg.Ack(false) // sukses
}

// processSingleConfig melakukan sinkronisasi transaksi secara batch menggunakan goroutine:
// fetch ERP secara paralel → simpan ke blockchain (status pending) → publish ke predict_requests queue.
func (w *SyncWorker) processSingleConfig(conf *domain.Configuration, erpURL, apiKey string) error {
	batchSize := 5
	type fetchResult struct {
		body []byte
		err  error
	}

	log.Printf("[Sync Worker] Memulai sinkronisasi transaksi batching (ukuran: %d) dari ERP: %s\n", batchSize, erpURL)

	ch := make(chan fetchResult, batchSize)
	for i := 0; i < batchSize; i++ {
		go func(index int) {
			log.Printf("[Sync Worker] [Batch %d] Fetching transaksi dari ERP...\n", index)
			req, err := http.NewRequest("GET", erpURL, nil)
			if err != nil {
				ch <- fetchResult{err: fmt.Errorf("gagal membuat request: %w", err)}
				return
			}
			if apiKey != "" {
				req.Header.Set("Authorization", "Bearer "+apiKey)
			}

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				ch <- fetchResult{err: fmt.Errorf("error fetch ERP: %w", err)}
				return
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				ch <- fetchResult{err: fmt.Errorf("error membaca body: %w", err)}
				return
			}
			log.Printf("[Sync Worker] [Batch %d] ✓ Berhasil fetch transaksi (%d bytes)\n", index, len(bodyBytes))
			ch <- fetchResult{body: bodyBytes}
		}(i)
	}

	var bodies [][]byte
	for i := 0; i < batchSize; i++ {
		res := <-ch
		if res.err != nil {
			log.Printf("[Sync Worker] Error saat batch fetch: %v\n", res.err)
			continue
		}
		bodies = append(bodies, res.body)
	}

	log.Printf("[Sync Worker] Berhasil mengambil total %d transaksi dari ERP. Merekam ke blockchain dan mengirim antrean prediksi...\n", len(bodies))

	for idx, bodyBytes := range bodies {
		var txInput map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &txInput); err != nil {
			log.Printf("[Sync Worker] [Tx %d] Gagal parse payload ERP: %v\n", idx, err)
			continue
		}

		vendorName, _ := txInput["vendor_name"].(string)
		amountIdrFloat, ok := txInput["amount_idr"].(float64)
		if !ok {
			log.Printf("[Sync Worker] [Tx %d] Payload ERP tidak valid (amount_idr kosong/invalid)\n", idx)
			continue
		}

		// Generate address hash dari nama vendor
		vendorHash := fmt.Sprintf("0x%x", sha256.Sum256([]byte(vendorName)))
		if len(vendorHash) > 42 {
			vendorHash = vendorHash[:42]
		}

		// Hash sistem berdasarkan UserID klien
		systemString := fmt.Sprintf("System_User_%d", conf.UserID)
		systemHashLong := fmt.Sprintf("0x%x", sha256.Sum256([]byte(systemString)))
		systemHash := systemHashLong[:42]

		txHashBytes := sha256.Sum256(bodyBytes)
		txHash := fmt.Sprintf("0x%x", txHashBytes)

		log.Printf("[Sync Worker] [Tx %d] Merekam transaksi ke blockchain dengan status 'pending'... TxHash: %s\n", idx, txHash)

		// 1. Simpan ke blockchain (status pending)
		evmTxHash, err := w.blockchainRepo.RecordTransaction(
			txHash,
			systemHash,
			vendorHash,
			amountIdrFloat,
			amountIdrFloat*0.0001,
			21000,
			string(bodyBytes),
		)
		if err != nil {
			log.Printf("[Sync Worker] [Tx %d] Gagal merekam transaksi ke blockchain: %v\n", idx, err)
			continue
		}

		log.Printf("[Sync Worker] [Tx %d] ✓ Berhasil direkam di blockchain (pending). EVM Hash: %s. Mengirim request prediksi...\n", idx, evmTxHash)

		// 2. Publish ke model service via AMQP
		err = w.broker.PublishPredictRequest(txHash, bodyBytes, ModelParams{
			VolumeSensitivity: conf.VolumeSensitivity,
			GeoThreshold:      conf.GeoThreshold,
			VelocityLimit:     conf.VelocityLimit,
		})
		if err != nil {
			log.Printf("[Sync Worker] [Tx %d] Gagal mengirim request prediksi: %v\n", idx, err)
			continue
		}
		log.Printf("[Sync Worker] [Tx %d] ✓ Request prediksi berhasil dipublish. TxHash: %s\n", idx, txHash)
	}

	return nil
}

// startPredictResponseConsumer memulai loop konsumsi hasil prediksi dari model_service.
func (w *SyncWorker) startPredictResponseConsumer() {
	msgs, err := w.broker.ConsumePredictResponses()
	if err != nil {
		log.Fatalf("[Sync Worker] Gagal memulai consumer predict_responses: %v\n", err)
	}

	log.Println("[Sync Worker] Consumer predict_responses aktif, menunggu hasil prediksi...")

	for msg := range msgs {
		w.handlePredictResponse(msg)
	}

	log.Println("[Sync Worker] Consumer predict_responses berhenti.")
}

type PredictResponsePayload struct {
	TxHash     string  `json:"tx_hash"`
	IsFraud    bool    `json:"is_fraud"`
	RiskScore  int     `json:"risk_score"`
	Verdict    string  `json:"verdict"`
	FlagReason string  `json:"flag_reason"`
	Error      string  `json:"error,omitempty"`
}

func (w *SyncWorker) handlePredictResponse(msg amqp.Delivery) {
	defer msg.Ack(false)

	var res PredictResponsePayload
	if err := json.Unmarshal(msg.Body, &res); err != nil {
		log.Printf("[Sync Worker] Gagal unmarshal hasil prediksi: %v\n", err)
		return
	}

	if res.TxHash == "" {
		log.Println("[Sync Worker] Hasil prediksi tidak memiliki tx_hash. Diabaikan.")
		return
	}

	if res.Error != "" {
		log.Printf("[Sync Worker] Hasil prediksi error untuk TxHash %s: %s. Mengupdate status di blockchain ke ERROR.\n", res.TxHash, res.Error)
		_, err := w.blockchainRepo.UpdateTransactionPrediction(
			res.TxHash,
			false,
			0,
			"ERROR",
			res.Error,
		)
		if err != nil {
			log.Printf("[Sync Worker] Gagal update status error transaksi ke blockchain: %v\n", err)
		}
		return
	}

	log.Printf("[Sync Worker] Memproses hasil prediksi untuk TxHash: %s (IsFraud: %t, Risk: %d, Verdict: %s)\n", res.TxHash, res.IsFraud, res.RiskScore, res.Verdict)

	evmHash, err := w.blockchainRepo.UpdateTransactionPrediction(
		res.TxHash,
		res.IsFraud,
		res.RiskScore,
		res.Verdict,
		res.FlagReason,
	)
	if err != nil {
		log.Printf("[Sync Worker] Gagal update hasil prediksi ke blockchain untuk TxHash %s: %v\n", res.TxHash, err)
		return
	}

	log.Printf("[Sync Worker] ✓ Hasil prediksi sukses diupdate di blockchain untuk TxHash: %s (EVM TxHash: %s)\n", res.TxHash, evmHash)
}

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
	blockRepo   repository.BlockRepository
	txRepo      repository.TransactionRepository
	settingRepo repository.SettingRepository
	broker      *Broker
}

func NewSyncWorker(
	b repository.BlockRepository,
	t repository.TransactionRepository,
	s repository.SettingRepository,
	broker *Broker,
) *SyncWorker {
	return &SyncWorker{
		blockRepo:   b,
		txRepo:      t,
		settingRepo: s,
		broker:      broker,
	}
}

// Start menjalankan dua goroutine:
//  1. Scheduler: setiap 30 menit publish SyncJob per konfigurasi ke queue.
//  2. Consumer: memproses SyncJob dari queue satu per satu.
func (w *SyncWorker) Start() {
	log.Println("[Worker] Memulai Background Sync Worker dengan RabbitMQ...")

	// Goroutine consumer — berjalan selamanya
	go w.startConsumer()

	// Goroutine scheduler — trigger pertama langsung, lalu setiap 30 menit
	go func() {
		w.publishAllJobs()
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			w.publishAllJobs()
		}
	}()
}

// publishAllJobs membaca semua konfigurasi dan mengirim SyncJob ke queue.
func (w *SyncWorker) publishAllJobs() {
	log.Println("[Worker] Mengambil konfigurasi ERP dari database...")

	confs, err := w.settingRepo.GetAllConfigurations()
	if err != nil {
		log.Printf("[Worker] Error mengambil konfigurasi: %v\n", err)
		return
	}

	if len(confs) == 0 {
		log.Println("[Worker] Tidak ada konfigurasi di database. Sync dilewati.")
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
			log.Printf("[Worker] Gagal publish SyncJob untuk UserID %d: %v\n", conf.UserID, err)
			continue
		}
		log.Printf("[Worker] SyncJob dipublish untuk UserID %d (endpoint: %s)\n", conf.UserID, endpoint)
	}
}

// startConsumer memulai loop konsumsi pesan dari RabbitMQ.
func (w *SyncWorker) startConsumer() {
	msgs, err := w.broker.Consume()
	if err != nil {
		log.Fatalf("[Worker] Gagal memulai consumer RabbitMQ: %v\n", err)
	}

	log.Println("[Worker] Consumer RabbitMQ aktif, menunggu SyncJob...")

	for msg := range msgs {
		w.handleMessage(msg)
	}

	log.Println("[Worker] Consumer RabbitMQ berhenti (channel ditutup).")
}

// handleMessage mendekode pesan dan memproses satu SyncJob.
func (w *SyncWorker) handleMessage(msg amqp.Delivery) {
	var job SyncJob
	if err := json.Unmarshal(msg.Body, &job); err != nil {
		log.Printf("[Worker] Pesan tidak valid, dibuang: %v\n", err)
		msg.Nack(false, false) // buang, jangan requeue
		return
	}

	log.Printf("[Worker] Memproses SyncJob — UserID: %d, Endpoint: %s\n", job.UserID, job.Endpoint)

	// Ambil konfigurasi lengkap dari DB (termasuk sensitivitas parameter model)
	conf, err := w.settingRepo.GetConfigurationByUserID(job.UserID)
	if err != nil {
		log.Printf("[Worker] Konfigurasi UserID %d tidak ditemukan: %v\n", job.UserID, err)
		msg.Nack(false, false)
		return
	}

	if err := w.processSingleConfig(conf, job.Endpoint, job.ApiKey); err != nil {
		log.Printf("[Worker] Gagal memproses SyncJob UserID %d: %v\n", job.UserID, err)
		msg.Nack(false, false) // gagal, jangan requeue (hindari loop)
		return
	}

	msg.Ack(false) // sukses
}

// processSingleConfig melakukan full pipeline untuk satu konfigurasi ERP:
// fetch ERP → call model service → simpan block & transaksi.
func (w *SyncWorker) processSingleConfig(conf *domain.Configuration, erpURL, apiKey string) error {
	// 1. Fetch dari ERP
	req, err := http.NewRequest("GET", erpURL, nil)
	if err != nil {
		return fmt.Errorf("gagal membuat request: %w", err)
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error fetch ERP %s: %w", erpURL, err)
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var txInput map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &txInput); err != nil {
		return fmt.Errorf("error parsing data ERP: %w", err)
	}

	vendorName, _ := txInput["vendor_name"].(string)
	amountIdrFloat, ok := txInput["amount_idr"].(float64)
	if !ok {
		return fmt.Errorf("payload ERP %s tidak valid (amount_idr kosong/invalid)", erpURL)
	}

	// 2. Kirim ke Model Service via AMQP RPC
	modelResult, err := w.broker.CallModelService(bodyBytes, ModelParams{
		VolumeSensitivity: conf.VolumeSensitivity,
		GeoThreshold:      conf.GeoThreshold,
		VelocityLimit:     conf.VelocityLimit,
	})
	if err != nil {
		return fmt.Errorf("error memanggil Model Service via AMQP: %w", err)
	}

	isFraud   := modelResult.IsFraud
	verdict   := modelResult.Verdict
	flagReason := modelResult.FlagReason
	riskScore := modelResult.RiskScore

	// Generate address hash dari nama vendor
	vendorHash := fmt.Sprintf("0x%x", sha256.Sum256([]byte(vendorName)))
	if len(vendorHash) > 42 {
		vendorHash = vendorHash[:42]
	}

	// Hash sistem berdasarkan UserID klien
	systemString := fmt.Sprintf("System_User_%d", conf.UserID)
	systemHashLong := fmt.Sprintf("0x%x", sha256.Sum256([]byte(systemString)))
	systemHash := systemHashLong[:42]

	txHash := fmt.Sprintf("0x%x", sha256.Sum256(bodyBytes))

	// 3. Buat Block & Transaksi
	latestHeight, _ := w.blockRepo.GetLatestBlockHeight()
	newHeight := latestHeight + 1

	parentHash := "0x0000000000000000000000000000000000000000000000000000000000000000"
	if newHeight > 1 {
		latestBlock, _ := w.blockRepo.GetBlockByHashOrHeight(fmt.Sprintf("%d", latestHeight))
		if latestBlock != nil {
			parentHash = latestBlock.Hash
		}
	}

	blockHashData := fmt.Sprintf("%d%s%d", newHeight, parentHash, time.Now().UnixNano())
	blockHash := fmt.Sprintf("0x%x", sha256.Sum256([]byte(blockHashData)))

	newBlock := &domain.Block{
		Height:           newHeight,
		Hash:             blockHash,
		ParentHash:       parentHash,
		Timestamp:        time.Now(),
		Size:             len(bodyBytes) + 200,
		Miner:            "0xTrustChainMiner01",
		TransactionCount: 1,
	}

	if err := w.blockRepo.CreateBlock(newBlock); err != nil {
		return fmt.Errorf("error membuat block: %w", err)
	}

	newTx := &domain.Transaction{
		Hash:        txHash,
		BlockHeight: newHeight,
		Status:      "success",
		FromAddress: systemHash,
		ToAddress:   vendorHash,
		Value:       amountIdrFloat,
		Fee:         amountIdrFloat * 0.0001,
		GasUsed:     21000,
		Timestamp:   time.Now(),
		IsFraud:     isFraud,
		Verdict:     verdict,
		FlagReason:  flagReason,
		RiskScore:   riskScore,
		Data:        string(bodyBytes),
	}

	if err := w.txRepo.CreateTransaction(newTx); err != nil {
		return fmt.Errorf("error membuat transaksi: %w", err)
	}

	log.Printf("[Worker] ✓ Block #%d berhasil disinkronisasi dari %s\n", newHeight, erpURL)
	return nil
}

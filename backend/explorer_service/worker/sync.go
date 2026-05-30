package worker

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"explorer_service/domain"
	"explorer_service/repository"
)

type SyncWorker struct {
	blockRepo   repository.BlockRepository
	txRepo      repository.TransactionRepository
	settingRepo repository.SettingRepository
	modelURL    string
}

func NewSyncWorker(b repository.BlockRepository, t repository.TransactionRepository, s repository.SettingRepository) *SyncWorker {
	modelURL := os.Getenv("MODEL_API_URL")
	if modelURL == "" {
		modelURL = "http://model_service:8000/predict" // internal docker network
	}

	return &SyncWorker{
		blockRepo:   b,
		txRepo:      t,
		settingRepo: s,
		modelURL:    modelURL,
	}
}

func (w *SyncWorker) Start() {
	log.Println("Starting Background Sync Worker...")
	ticker := time.NewTicker(30 * time.Second)

	go func() {
		for {
			<-ticker.C
			w.syncData()
		}
	}()

	// Run once immediately
	w.syncData()
}

func (w *SyncWorker) syncData() {
	log.Println("[Worker] Fetching ERP endpoints from configurations...")

	confs, err := w.settingRepo.GetAllConfigurations()
	if err != nil {
		log.Printf("[Worker] Error fetching configurations: %v\n", err)
		return
	}

	if len(confs) == 0 {
		log.Println("[Worker] No configurations found in database. Skipping sync.")
		return
	}

	for _, conf := range confs {
		erpURL := strings.TrimSpace(conf.Endpoint)
		if erpURL == "" {
			continue // Skip if no endpoint is configured
		}

		log.Printf("[Worker] Syncing data from ERP: %s (User ID: %d)\n", erpURL, conf.UserID)

		// 1. Fetch from ERP
		req, _ := http.NewRequest("GET", erpURL, nil)
		if conf.ApiKey != "" {
			req.Header.Set("Authorization", "Bearer "+conf.ApiKey) // Optional Auth
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[Worker] Error fetching from ERP %s: %v\n", erpURL, err)
			continue
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var txInput map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &txInput); err != nil {
			log.Printf("[Worker] Error parsing ERP data from %s: %v\n", erpURL, err)
			continue
		}

		vendorName, _ := txInput["vendor_name"].(string)
		amountIdrFloat, ok := txInput["amount_idr"].(float64)
		if !ok {
			log.Printf("[Worker] Invalid payload from %s (amount_idr missing or invalid)\n", erpURL)
			continue
		}

		// 2. Send to Model Service
		modelURLWithParams := fmt.Sprintf("%s?volume_sensitivity=%d&geo_threshold=%d&velocity_limit=%d", 
			w.modelURL, conf.VolumeSensitivity, conf.GeoThreshold, conf.VelocityLimit)

		reqBody := bytes.NewBuffer(bodyBytes)
		modelResp, err := http.Post(modelURLWithParams, "application/json", reqBody)
		if err != nil {
			log.Printf("[Worker] Error calling Model Service: %v\n", err)
			continue
		}

		var modelOutput map[string]interface{}
		err = json.NewDecoder(modelResp.Body).Decode(&modelOutput)
		modelResp.Body.Close()
		if err != nil {
			log.Printf("[Worker] Error parsing Model response: %v\n", err)
			continue
		}

		isFraud, _ := modelOutput["is_fraud"].(bool)
		verdict, _ := modelOutput["verdict"].(string)
		flagReason, _ := modelOutput["flag_reason"].(string)
		riskScoreFloat, _ := modelOutput["risk_score"].(float64)
		riskScore := int(riskScoreFloat)

		// Generate address for vendor (Hash vendor name)
		vendorHash := fmt.Sprintf("0x%x", sha256.Sum256([]byte(vendorName)))
		if len(vendorHash) > 42 {
			vendorHash = vendorHash[:42]
		}

		// System Hash indicates the Client's ERP system
		systemString := fmt.Sprintf("System_User_%d", conf.UserID)
		systemHashLong := fmt.Sprintf("0x%x", sha256.Sum256([]byte(systemString)))
		systemHash := systemHashLong[:42]

		txHash := fmt.Sprintf("0x%x", sha256.Sum256(bodyBytes))

		// 3. Create Block & Transaction
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
			Size:             len(bodyBytes) + 200, // mock size
			Miner:            "0xTrustChainMiner01",
			TransactionCount: 1,
		}

		if err := w.blockRepo.CreateBlock(newBlock); err != nil {
			log.Printf("[Worker] Error creating block: %v\n", err)
			continue
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
			log.Printf("[Worker] Error creating transaction: %v\n", err)
			continue
		}

		log.Printf("[Worker] Successfully synced block #%d with 1 transaction from %s\n", newHeight, erpURL)
	}
}

package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	syncQueueName    = "sync_jobs"
	predictReqQueue  = "predict_requests"
	predictRespQueue = "predict_responses"
)

// SyncJob adalah payload pesan untuk pekerjaan sinkronisasi ERP.
type SyncJob struct {
	UserID   uint   `json:"user_id"`
	ConfigID uint   `json:"config_id"`
	Endpoint string `json:"endpoint"`
	ApiKey   string `json:"api_key"`
}

// ModelParams berisi parameter tuning yang dikirim ke model_service.
type ModelParams struct {
	VolumeSensitivity int `json:"volume_sensitivity"`
	GeoThreshold      int `json:"geo_threshold"`
	VelocityLimit     int `json:"velocity_limit"`
}

// ModelResponse adalah hasil prediksi dari model_service melalui AMQP.
type ModelResponse struct {
	VendorName    string  `json:"vendor_name"`
	AmountIdr     float64 `json:"amount_idr"`
	RiskScore     int     `json:"risk_score"`
	IFScore       float64 `json:"if_score"`
	LSTMProb      float64 `json:"lstm_prob"`
	EnsembleScore float64 `json:"ensemble_score"`
	IsFraud       bool    `json:"is_fraud"`
	Verdict       string  `json:"verdict"`
	FlagReason    string  `json:"flag_reason"`
	Error         string  `json:"error,omitempty"`
}

// Broker mengelola koneksi dan channel ke RabbitMQ.
type Broker struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewBroker membuat koneksi baru ke RabbitMQ dengan mekanisme retry.
func NewBroker(amqpURL string) (*Broker, error) {
	var conn *amqp.Connection
	var err error

	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(amqpURL)
		if err == nil {
			break
		}
		log.Printf("[Broker] Menunggu RabbitMQ... percobaan %d/10: %v\n", i+1, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Deklarasi queue sync_jobs
	if _, err = ch.QueueDeclare(syncQueueName, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("gagal deklarasi %s: %w", syncQueueName, err)
	}

	// Deklarasi queue predict_requests
	if _, err = ch.QueueDeclare(predictReqQueue, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("gagal deklarasi %s: %w", predictReqQueue, err)
	}

	// Deklarasi queue predict_responses
	if _, err = ch.QueueDeclare(predictRespQueue, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("gagal deklarasi %s: %w", predictRespQueue, err)
	}

	log.Printf("[Broker] Terhubung ke RabbitMQ. Queue '%s', '%s', dan '%s' siap.\n", syncQueueName, predictReqQueue, predictRespQueue)
	return &Broker{conn: conn, channel: ch}, nil
}

// Publish mengirim SyncJob ke queue sync_jobs.
func (b *Broker) Publish(job SyncJob) error {
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}
	log.Printf("[Broker] Publishing SyncJob untuk UserID %d ke queue '%s'\n", job.UserID, syncQueueName)
	return b.channel.Publish("", syncQueueName, false, false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
}

// Consume memulai konsumsi pesan dari sync_jobs queue.
func (b *Broker) Consume() (<-chan amqp.Delivery, error) {
	if err := b.channel.Qos(1, 0, false); err != nil {
		return nil, err
	}
	return b.channel.Consume(syncQueueName, "", false, false, false, false, nil)
}

// PublishPredictRequest mengirim payload transaksi ke model_service secara asynchronous.
func (b *Broker) PublishPredictRequest(txHash string, txData []byte, params ModelParams) error {
	var payloadMap map[string]interface{}
	if err := json.Unmarshal(txData, &payloadMap); err != nil {
		return err
	}
	payloadMap["tx_hash"] = txHash
	payloadMap["_params"] = params

	enrichedBody, err := json.Marshal(payloadMap)
	if err != nil {
		return err
	}

	log.Printf("[Broker] Publishing prediction request untuk TxHash: %s ke queue '%s'\n", txHash, predictReqQueue)
	return b.channel.Publish("", predictReqQueue, false, false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         enrichedBody,
		},
	)
}

// ConsumePredictResponses memulai konsumsi pesan dari predict_responses queue.
func (b *Broker) ConsumePredictResponses() (<-chan amqp.Delivery, error) {
	if err := b.channel.Qos(1, 0, false); err != nil {
		return nil, err
	}
	return b.channel.Consume(predictRespQueue, "", false, false, false, false, nil)
}

// CallModelService mengirim payload ke model_service via AMQP RPC (request/reply pattern).
// Catatan: Ini dipertahankan untuk kompatibilitas, sync worker sekarang menggunakan PublishPredictRequest.
func (b *Broker) CallModelService(payload []byte, params ModelParams) (*ModelResponse, error) {
	replyQueue, err := b.channel.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat reply queue: %w", err)
	}

	replies, err := b.channel.Consume(replyQueue.Name, "", true, true, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal subscribe reply queue: %w", err)
	}

	corrID := fmt.Sprintf("corr-%d", time.Now().UnixNano())

	var payloadMap map[string]interface{}
	if err := json.Unmarshal(payload, &payloadMap); err != nil {
		return nil, fmt.Errorf("gagal parse payload: %w", err)
	}
	payloadMap["_params"] = params
	enrichedBody, err := json.Marshal(payloadMap)
	if err != nil {
		return nil, fmt.Errorf("gagal marshal enriched payload: %w", err)
	}

	if err := b.channel.Publish("", predictReqQueue, false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       replyQueue.Name,
			Body:          enrichedBody,
		},
	); err != nil {
		return nil, fmt.Errorf("gagal publish ke predict_requests: %w", err)
	}

	timeout := time.After(30 * time.Second)
	for {
		select {
		case msg, ok := <-replies:
			if !ok {
				return nil, fmt.Errorf("reply channel ditutup sebelum menerima response")
			}
			if msg.CorrelationId != corrID {
				continue
			}
			var result ModelResponse
			if err := json.Unmarshal(msg.Body, &result); err != nil {
				return nil, fmt.Errorf("gagal parse model response: %w", err)
			}
			if result.Error != "" {
				return nil, fmt.Errorf("model service error: %s", result.Error)
			}
			return &result, nil

		case <-timeout:
			return nil, fmt.Errorf("timeout menunggu response dari model_service (30s)")
		}
	}
}

// Close menutup channel dan koneksi RabbitMQ.
func (b *Broker) Close() {
	if b.channel != nil {
		b.channel.Close()
	}
	if b.conn != nil {
		b.conn.Close()
	}
}

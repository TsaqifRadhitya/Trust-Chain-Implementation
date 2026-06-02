package domain

import "time"

type Transaction struct {
	Hash        string    `json:"hash"`
	BlockHeight int       `json:"block_height"`
	Status      string    `json:"status"` // e.g., "success"
	FromAddress string    `json:"from"`
	ToAddress   string    `json:"to"`
	Value       float64   `json:"value"`
	Fee         float64   `json:"fee"`
	GasUsed     int       `json:"gas_used"`
	Timestamp   time.Time `json:"timestamp"`
	IsFraud     bool      `json:"is_fraud"`
	Verdict     string    `json:"verdict"`
	FlagReason  string    `json:"flag_reason"`
	RiskScore   int       `json:"risk_score"`
	Data        string    `json:"data"` // JSON string of original TransactionInput
}

package domain

import "time"

type Block struct {
	Height           int           `json:"height"`
	Hash             string        `json:"hash"`
	ParentHash       string        `json:"parent_hash"`
	Timestamp        time.Time     `json:"timestamp"`
	Size             int           `json:"size"`
	Miner            string        `json:"miner"`
	TransactionCount int           `json:"tx_count"`
	Transactions     []Transaction `json:"transactions"`
}

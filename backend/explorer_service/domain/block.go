package domain

import "time"

type Block struct {
	Height           int           `json:"height" gorm:"primaryKey;autoIncrement:false"`
	Hash             string        `json:"hash" gorm:"unique;not null"`
	ParentHash       string        `json:"parent_hash" gorm:"not null"`
	Timestamp        time.Time     `json:"timestamp" gorm:"not null"`
	Size             int           `json:"size"`
	Miner            string        `json:"miner"`
	TransactionCount int           `json:"tx_count"`
	Transactions     []Transaction `json:"transactions" gorm:"foreignKey:BlockHeight;references:Height"`
}

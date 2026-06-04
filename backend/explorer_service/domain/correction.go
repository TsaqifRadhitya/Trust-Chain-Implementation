package domain

import "time"

type Correction struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	TxHash       string    `json:"tx_hash" gorm:"uniqueIndex;not null"`
	IsCorrected  bool      `json:"is_corrected" gorm:"default:false"`
	ActualStatus string    `json:"actual_status"` // e.g., "Safe", "Fraud"
	Reason       string    `json:"reason"`
	CorrectedBy  string    `json:"corrected_by"`
	UpdatedAt    time.Time `json:"updated_at"`
}

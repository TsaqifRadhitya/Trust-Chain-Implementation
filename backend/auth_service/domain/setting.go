package domain

import (
	"context"
	"time"
)

type Configuration struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	UserID            uint      `json:"user_id" gorm:"uniqueIndex;not null"`
	ErpType           string    `json:"erp_type"`
	Endpoint          string    `json:"endpoint"`
	ApiKey            string    `json:"api_key"`
	VolumeSensitivity int       `json:"volume_sensitivity"`
	GeoThreshold      int       `json:"geo_threshold"`
	VelocityLimit     int       `json:"velocity_limit"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type SettingRepository interface {
	GetByUserID(ctx context.Context, userID uint) (*Configuration, error)
	Upsert(ctx context.Context, config *Configuration) error
}

type SettingUsecase interface {
	GetSettings(ctx context.Context, userID uint) (*Configuration, error)
	UpdateSettings(ctx context.Context, userID uint, config *Configuration) (*Configuration, error)
}

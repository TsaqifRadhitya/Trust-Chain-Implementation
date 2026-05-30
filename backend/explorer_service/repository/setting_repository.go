package repository

import (
	"explorer_service/domain"

	"gorm.io/gorm"
)

type SettingRepository interface {
	GetAllConfigurations() ([]domain.Configuration, error)
}

type settingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) SettingRepository {
	return &settingRepository{db: db}
}

func (r *settingRepository) GetAllConfigurations() ([]domain.Configuration, error) {
	var confs []domain.Configuration
	err := r.db.Find(&confs).Error
	return confs, err
}

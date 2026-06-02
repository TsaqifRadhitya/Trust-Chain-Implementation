package repository

import (
	"explorer_service/domain"

	"gorm.io/gorm"
)

type SettingRepository interface {
	GetAllConfigurations() ([]domain.Configuration, error)
	GetConfigurationByUserID(userID uint) (*domain.Configuration, error)
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

func (r *settingRepository) GetConfigurationByUserID(userID uint) (*domain.Configuration, error) {
	var conf domain.Configuration
	err := r.db.Where("user_id = ?", userID).First(&conf).Error
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

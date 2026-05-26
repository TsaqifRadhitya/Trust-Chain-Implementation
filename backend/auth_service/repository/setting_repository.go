package repository

import (
	"context"
	"backend/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type settingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) domain.SettingRepository {
	return &settingRepository{db: db}
}

func (r *settingRepository) GetByUserID(ctx context.Context, userID uint) (*domain.Configuration, error) {
	var conf domain.Configuration
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&conf).Error
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

func (r *settingRepository) Upsert(ctx context.Context, config *domain.Configuration) error {
	// Melakukan Upsert jika user_id sudah ada
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"erp_type", "endpoint", "api_key", "volume_sensitivity", "geo_threshold", "velocity_limit"}),
	}).Save(config).Error
}

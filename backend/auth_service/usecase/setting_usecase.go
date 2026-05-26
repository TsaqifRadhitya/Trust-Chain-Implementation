package usecase

import (
	"context"
	"backend/domain"
)

type settingUsecase struct {
	settingRepo domain.SettingRepository
}

func NewSettingUsecase(settingRepo domain.SettingRepository) domain.SettingUsecase {
	return &settingUsecase{settingRepo: settingRepo}
}

func (u *settingUsecase) GetSettings(ctx context.Context, userID uint) (*domain.Configuration, error) {
	config, err := u.settingRepo.GetByUserID(ctx, userID)
	if err != nil {
		// Jika belum ada, return default konfigurasi kosong
		return &domain.Configuration{
			UserID:            userID,
			ErpType:           "SAP S/4HANA",
			Endpoint:          "https://erp.internal.company.com/api/v2/transactions",
			ApiKey:            "sk-trustchain-default-placeholder",
			VolumeSensitivity: 85,
			GeoThreshold:      50,
			VelocityLimit:     70,
		}, nil
	}
	return config, nil
}

func (u *settingUsecase) UpdateSettings(ctx context.Context, userID uint, config *domain.Configuration) (*domain.Configuration, error) {
	config.UserID = userID
	err := u.settingRepo.Upsert(ctx, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

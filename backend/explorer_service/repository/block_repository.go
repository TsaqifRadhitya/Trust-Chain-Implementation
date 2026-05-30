package repository

import (
	"explorer_service/domain"

	"gorm.io/gorm"
)

type BlockRepository interface {
	CreateBlock(block *domain.Block) error
	GetRecentBlocks(limit int, offset int) ([]domain.Block, error)
	GetBlockByHashOrHeight(hashOrHeight string) (*domain.Block, error)
	GetLatestBlockHeight() (int, error)
}

type blockRepository struct {
	db *gorm.DB
}

func NewBlockRepository(db *gorm.DB) BlockRepository {
	return &blockRepository{db: db}
}

func (r *blockRepository) CreateBlock(block *domain.Block) error {
	return r.db.Create(block).Error
}

func (r *blockRepository) GetRecentBlocks(limit int, offset int) ([]domain.Block, error) {
	var blocks []domain.Block
	err := r.db.Order("height desc").Limit(limit).Offset(offset).Find(&blocks).Error
	return blocks, err
}

func (r *blockRepository) GetBlockByHashOrHeight(hashOrHeight string) (*domain.Block, error) {
	var block domain.Block
	err := r.db.Preload("Transactions").Where("hash = ? OR cast(height as varchar) = ?", hashOrHeight, hashOrHeight).First(&block).Error
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (r *blockRepository) GetLatestBlockHeight() (int, error) {
	var block domain.Block
	err := r.db.Order("height desc").First(&block).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return block.Height, nil
}

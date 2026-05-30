package repository

import (
	"explorer_service/domain"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	CreateTransaction(tx *domain.Transaction) error
	GetRecentTransactions(limit int) ([]domain.Transaction, error)
	GetTransactionByHash(hash string) (*domain.Transaction, error)
	GetTransactionsByAddress(address string) ([]domain.Transaction, error)
	GetBalanceByAddress(address string) (float64, error)
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) CreateTransaction(tx *domain.Transaction) error {
	return r.db.Create(tx).Error
}

func (r *transactionRepository) GetRecentTransactions(limit int) ([]domain.Transaction, error) {
	var txs []domain.Transaction
	err := r.db.Order("timestamp desc").Limit(limit).Find(&txs).Error
	return txs, err
}

func (r *transactionRepository) GetTransactionByHash(hash string) (*domain.Transaction, error) {
	var tx domain.Transaction
	err := r.db.Where("hash = ?", hash).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *transactionRepository) GetTransactionsByAddress(address string) ([]domain.Transaction, error) {
	var txs []domain.Transaction
	err := r.db.Where("from_address = ? OR to_address = ?", address, address).Order("timestamp desc").Find(&txs).Error
	return txs, err
}

func (r *transactionRepository) GetBalanceByAddress(address string) (float64, error) {
	var totalReceived float64
	var totalSent float64

	r.db.Model(&domain.Transaction{}).Where("to_address = ?", address).Select("COALESCE(sum(value), 0)").Scan(&totalReceived)
	r.db.Model(&domain.Transaction{}).Where("from_address = ?", address).Select("COALESCE(sum(value + fee), 0)").Scan(&totalSent)

	return totalReceived - totalSent, nil
}

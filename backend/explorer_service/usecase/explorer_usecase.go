package usecase

import (
	"fmt"
	"strconv"
	"time"

	"explorer_service/domain"
	"explorer_service/repository"
)

type ValidationDetail struct {
	Height            int       `json:"height"`
	Hash              string    `json:"hash"`
	ParentHash        string    `json:"parent_hash"`
	PreviousBlockHash string    `json:"previous_block_hash"`
	Timestamp         time.Time `json:"timestamp"`
	TxCount           int       `json:"tx_count"`
	Status            string    `json:"status"` // "OK" or "CORRUPTED"
	Error             string    `json:"error,omitempty"`
}

type ChainValidationResult struct {
	IsValid         bool               `json:"is_valid"`
	TotalBlocks     int                `json:"total_blocks"`
	ValidatedBlocks int                `json:"validated_blocks"`
	Details         []ValidationDetail `json:"details"`
}

type ExplorerUsecase interface {
	GetRecentBlocks(limit int, page int) ([]domain.Block, error)
	GetBlockDetail(hashOrHeight string) (*domain.Block, error)
	GetRecentTransactions(limit int) ([]domain.Transaction, error)
	GetTransactionDetail(hash string) (*domain.Transaction, error)
	GetAddressDetail(address string) (map[string]interface{}, error)
	Search(query string) (map[string]interface{}, error)
	ValidateChain() (*ChainValidationResult, error)
	GetStats() (map[string]interface{}, error)
	AddCorrection(txHash string, actualStatus string, reason string, correctedBy string) (*domain.Correction, error)
}

type explorerUsecase struct {
	blockchainRepo repository.BlockchainRepository
}

func NewExplorerUsecase(b repository.BlockchainRepository) ExplorerUsecase {
	return &explorerUsecase{
		blockchainRepo: b,
	}
}

func (u *explorerUsecase) GetRecentBlocks(limit int, page int) ([]domain.Block, error) {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	return u.blockchainRepo.GetRecentBlocks(limit, page)
}

func (u *explorerUsecase) GetBlockDetail(hashOrHeight string) (*domain.Block, error) {
	return u.blockchainRepo.GetBlockByHashOrHeight(hashOrHeight)
}

func (u *explorerUsecase) GetRecentTransactions(limit int) ([]domain.Transaction, error) {
	if limit <= 0 {
		limit = 10
	}
	return u.blockchainRepo.GetRecentTransactions(limit)
}

func (u *explorerUsecase) GetTransactionDetail(hash string) (*domain.Transaction, error) {
	return u.blockchainRepo.GetTransactionByHash(hash)
}

func (u *explorerUsecase) GetAddressDetail(address string) (map[string]interface{}, error) {
	balance, err := u.blockchainRepo.GetBalanceByAddress(address)
	if err != nil {
		return nil, err
	}

	txs, err := u.blockchainRepo.GetTransactionsByAddress(address)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"address":      address,
		"balance":      balance,
		"tx_count":     len(txs),
		"transactions": txs,
	}, nil
}

func (u *explorerUsecase) Search(query string) (map[string]interface{}, error) {
	tx, err := u.blockchainRepo.GetTransactionByHash(query)
	if err == nil && tx != nil {
		return map[string]interface{}{
			"type":       "transaction",
			"hash_or_id": tx.Hash,
		}, nil
	}

	block, err := u.blockchainRepo.GetBlockByHashOrHeight(query)
	if err == nil && block != nil {
		return map[string]interface{}{
			"type":       "block",
			"hash_or_id": block.Hash,
		}, nil
	}

	txs, err := u.blockchainRepo.GetTransactionsByAddress(query)
	if err == nil && len(txs) > 0 {
		return map[string]interface{}{
			"type":       "address",
			"hash_or_id": query,
		}, nil
	}

	return nil, nil
}

func (u *explorerUsecase) ValidateChain() (*ChainValidationResult, error) {
	latest, err := u.blockchainRepo.GetLatestBlockHeight()
	if err != nil {
		return nil, err
	}

	result := &ChainValidationResult{
		IsValid:     true,
		TotalBlocks: latest,
		Details:     make([]ValidationDetail, 0),
	}

	var prevHash string
	for i := 1; i <= latest; i++ {
		block, err := u.blockchainRepo.GetBlockHeaderOnly(strconv.Itoa(i))
		detail := ValidationDetail{
			Height: i,
		}

		if err != nil {
			detail.Status = "CORRUPTED"
			detail.Error = fmt.Sprintf("Failed to fetch block: %v", err)
			result.IsValid = false
			result.Details = append(result.Details, detail)
			continue
		}

		detail.Hash = block.Hash
		detail.ParentHash = block.ParentHash
		detail.Timestamp = block.Timestamp
		detail.TxCount = block.TransactionCount
		detail.PreviousBlockHash = prevHash

		// Validate block parent hash link
		if i > 1 {
			if block.ParentHash != prevHash {
				detail.Status = "CORRUPTED"
				detail.Error = fmt.Sprintf("Hash chain broken: ParentHash (%s) does not match previous block hash (%s)", block.ParentHash, prevHash)
				result.IsValid = false
			} else {
				detail.Status = "OK"
			}
		} else {
			// Genesis block verification
			detail.Status = "OK"
		}

		result.ValidatedBlocks++
		result.Details = append(result.Details, detail)
		prevHash = block.Hash
	}

	return result, nil
}

func (u *explorerUsecase) GetStats() (map[string]interface{}, error) {
	total, fraud, err := u.blockchainRepo.GetStats()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"total_transactions": total,
		"total_anomalies":    fraud,
	}, nil
}
func (u *explorerUsecase) AddCorrection(txHash string, actualStatus string, reason string, correctedBy string) (*domain.Correction, error) {
	return u.blockchainRepo.AddCorrection(txHash, actualStatus, reason, correctedBy)
}


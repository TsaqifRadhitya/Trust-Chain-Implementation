package usecase

import (
	"explorer_service/domain"
	"explorer_service/repository"
)

type ExplorerUsecase interface {
	GetRecentBlocks(limit int, page int) ([]domain.Block, error)
	GetBlockDetail(hashOrHeight string) (*domain.Block, error)
	GetRecentTransactions(limit int) ([]domain.Transaction, error)
	GetTransactionDetail(hash string) (*domain.Transaction, error)
	GetAddressDetail(address string) (map[string]interface{}, error)
	Search(query string) (map[string]interface{}, error)
}

type explorerUsecase struct {
	blockRepo repository.BlockRepository
	txRepo    repository.TransactionRepository
}

func NewExplorerUsecase(b repository.BlockRepository, t repository.TransactionRepository) ExplorerUsecase {
	return &explorerUsecase{
		blockRepo: b,
		txRepo:    t,
	}
}

func (u *explorerUsecase) GetRecentBlocks(limit int, page int) ([]domain.Block, error) {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	return u.blockRepo.GetRecentBlocks(limit, offset)
}

func (u *explorerUsecase) GetBlockDetail(hashOrHeight string) (*domain.Block, error) {
	return u.blockRepo.GetBlockByHashOrHeight(hashOrHeight)
}

func (u *explorerUsecase) GetRecentTransactions(limit int) ([]domain.Transaction, error) {
	if limit <= 0 {
		limit = 10
	}
	return u.txRepo.GetRecentTransactions(limit)
}

func (u *explorerUsecase) GetTransactionDetail(hash string) (*domain.Transaction, error) {
	return u.txRepo.GetTransactionByHash(hash)
}

func (u *explorerUsecase) GetAddressDetail(address string) (map[string]interface{}, error) {
	balance, err := u.txRepo.GetBalanceByAddress(address)
	if err != nil {
		return nil, err
	}

	txs, err := u.txRepo.GetTransactionsByAddress(address)
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
	tx, err := u.txRepo.GetTransactionByHash(query)
	if err == nil && tx != nil {
		return map[string]interface{}{
			"type":       "transaction",
			"hash_or_id": tx.Hash,
		}, nil
	}

	block, err := u.blockRepo.GetBlockByHashOrHeight(query)
	if err == nil && block != nil {
		return map[string]interface{}{
			"type":       "block",
			"hash_or_id": block.Hash,
		}, nil
	}

	txs, err := u.txRepo.GetTransactionsByAddress(query)
	if err == nil && len(txs) > 0 {
		return map[string]interface{}{
			"type":       "address",
			"hash_or_id": query,
		}, nil
	}

	return nil, nil
}

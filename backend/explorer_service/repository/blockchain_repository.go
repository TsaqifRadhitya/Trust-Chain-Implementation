package repository

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"explorer_service/contract"
	"explorer_service/domain"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type BlockchainRepository interface {
	GetLatestBlockHeight() (int, error)
	GetRecentBlocks(limit int, page int) ([]domain.Block, error)
	GetBlockByHashOrHeight(hashOrHeight string) (*domain.Block, error)
	GetTransactionByHash(hash string) (*domain.Transaction, error)
	GetRecentTransactions(limit int) ([]domain.Transaction, error)
	GetBalanceByAddress(address string) (float64, error)
	GetTransactionsByAddress(address string) ([]domain.Transaction, error)
	RecordTransaction(txHash string, fromAddr string, toAddr string, value float64, fee float64, gasUsed int, data string) (string, error)
	UpdateTransactionPrediction(txHash string, isFraud bool, riskScore int, verdict string, flagReason string) (string, error)
	GetBlockHeaderOnly(hashOrHeight string) (*domain.Block, error)
	GetStats() (int, int, error)
}

type blockchainRepository struct {
	client          *ethclient.Client
	privateKey      *ecdsa.PrivateKey
	senderAddress   common.Address
	contractAddress common.Address
	mu              sync.Mutex
}

func NewBlockchainRepository(ganacheURL string, privateKeyHex string) (BlockchainRepository, error) {
	log.Printf("[Blockchain Repository] Connecting to Ganache at %s...\n", ganacheURL)
	client, err := ethclient.Dial(ganacheURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ganache: %w", err)
	}

	privKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey := privKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key")
	}

	senderAddr := crypto.PubkeyToAddress(*publicKeyECDSA)
	log.Printf("[Blockchain Repository] Loaded private key. Sender address: %s\n", senderAddr.Hex())

	repo := &blockchainRepository{
		client:        client,
		privateKey:    privKey,
		senderAddress: senderAddr,
	}

	// Deploy smart contract at startup
	log.Println("[Blockchain Repository] Deploying TrustChain smart contract...")
	contractAddr, err := deployContract(client, privKey, senderAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy contract: %w", err)
	}
	repo.contractAddress = contractAddr
	log.Printf("[Blockchain Repository] ✓ TrustChain contract deployed successfully at: %s\n", contractAddr.Hex())

	return repo, nil
}

func deployContract(client *ethclient.Client, privateKey *ecdsa.PrivateKey, senderAddress common.Address) (common.Address, error) {
	nonce, err := client.PendingNonceAt(context.Background(), senderAddress)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasLimit := uint64(3000000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to suggest gas price: %w", err)
	}

	bytecodeBytes := common.FromHex(contract.TrustChainBytecode)
	tx := types.NewContractCreation(nonce, big.NewInt(0), gasLimit, gasPrice, bytecodeBytes)

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to get chain ID: %w", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to sign deploy transaction: %w", err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to send deploy transaction: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	for {
		receipt, err := client.TransactionReceipt(ctx, signedTx.Hash())
		if err == nil {
			return receipt.ContractAddress, nil
		}
		select {
		case <-ctx.Done():
			return common.Address{}, fmt.Errorf("timeout waiting for contract deploy to be mined")
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func (r *blockchainRepository) callContractWrite(methodName string, args ...interface{}) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Printf("[Blockchain Repository] Writing to contract: %s\n", methodName)
	parsedABI, err := abi.JSON(strings.NewReader(contract.TrustChainABI))
	if err != nil {
		return "", err
	}

	input, err := parsedABI.Pack(methodName, args...)
	if err != nil {
		return "", fmt.Errorf("failed to pack arguments for %s: %w", methodName, err)
	}

	nonce, err := r.client.PendingNonceAt(context.Background(), r.senderAddress)
	if err != nil {
		return "", err
	}

	gasLimit := uint64(3000000)
	gasPrice, err := r.client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &r.contractAddress,
		Value:    big.NewInt(0),
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     input,
	})

	chainID, err := r.client.ChainID(context.Background())
	if err != nil {
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), r.privateKey)
	if err != nil {
		return "", err
	}

	err = r.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	log.Printf("[Blockchain Repository] Sent transaction for %s. EVM TxHash: %s. Waiting for it to be mined...\n", methodName, signedTx.Hash().Hex())

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	for {
		_, isPending, err := r.client.TransactionByHash(ctx, signedTx.Hash())
		if err != nil {
			return "", err
		}
		if !isPending {
			break
		}
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("timeout waiting for transaction to be mined")
		case <-time.After(500 * time.Millisecond):
		}
	}

	log.Printf("[Blockchain Repository] ✓ Contract call %s mined. EVM TxHash: %s\n", methodName, signedTx.Hash().Hex())
	return signedTx.Hash().Hex(), nil
}

func (r *blockchainRepository) callContractRead(methodName string, args ...interface{}) ([]interface{}, error) {
	parsedABI, err := abi.JSON(strings.NewReader(contract.TrustChainABI))
	if err != nil {
		return nil, err
	}

	input, err := parsedABI.Pack(methodName, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack arguments for %s: %w", methodName, err)
	}

	msg := ethereum.CallMsg{
		To:   &r.contractAddress,
		Data: input,
	}

	output, err := r.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, err
	}

	results, err := parsedABI.Unpack(methodName, output)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack results for %s: %w", methodName, err)
	}

	return results, nil
}

func (r *blockchainRepository) GetLatestBlockHeight() (int, error) {
	header, err := r.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	return int(header.Number.Uint64()), nil
}

func (r *blockchainRepository) GetRecentBlocks(limit int, page int) ([]domain.Block, error) {
	latest, err := r.GetLatestBlockHeight()
	if err != nil {
		return nil, err
	}

	if latest == 0 {
		return []domain.Block{}, nil
	}

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit
	start := latest - offset
	if start < 1 {
		return []domain.Block{}, nil
	}

	end := start - limit + 1
	if end < 1 {
		end = 1
	}

	log.Printf("[Blockchain Repository] Fetching recent blocks from height %d to %d\n", start, end)

	var blocks []domain.Block
	for i := start; i >= end; i-- {
		b, err := r.GetBlockByHashOrHeight(strconv.Itoa(i))
		if err == nil && b != nil {
			blocks = append(blocks, *b)
		}
	}

	return blocks, nil
}

func (r *blockchainRepository) GetBlockByHashOrHeight(hashOrHeight string) (*domain.Block, error) {
	var block *types.Block
	var err error

	if height, errHeight := strconv.Atoi(hashOrHeight); errHeight == nil {
		block, err = r.client.BlockByNumber(context.Background(), big.NewInt(int64(height)))
	} else {
		block, err = r.client.BlockByHash(context.Background(), common.HexToHash(hashOrHeight))
	}

	if err != nil {
		return nil, err
	}

	blockNum := block.Number()
	txCountRes, err := r.callContractRead("getBlockTransactionCount", blockNum)
	if err != nil {
		return nil, fmt.Errorf("failed to get block transaction count: %w", err)
	}
	txCount := txCountRes[0].(*big.Int).Int64()

	log.Printf("[Blockchain Repository] Block %d has %d transactions in TrustChain contract\n", blockNum.Uint64(), txCount)

	var domainTxs []domain.Transaction
	for i := int64(0); i < txCount; i++ {
		hashRes, err := r.callContractRead("getBlockTransactionHashAtIndex", blockNum, big.NewInt(i))
		if err == nil {
			txHash := hashRes[0].(string)
			tx, err := r.GetTransactionByHash(txHash)
			if err == nil && tx != nil {
				domainTxs = append(domainTxs, *tx)
			}
		}
	}

	return &domain.Block{
		Height:           int(block.NumberU64()),
		Hash:             block.Hash().Hex(),
		ParentHash:       block.ParentHash().Hex(),
		Timestamp:        time.Unix(int64(block.Time()), 0),
		Size:             int(block.Size()),
		Miner:            block.Coinbase().Hex(),
		TransactionCount: int(txCount),
		Transactions:     domainTxs,
	}, nil
}

func (r *blockchainRepository) GetTransactionByHash(hash string) (*domain.Transaction, error) {
	baseRes, err := r.callContractRead("getTransactionBase", hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction base for hash %s: %w", hash, err)
	}

	predRes, err := r.callContractRead("getTransactionPrediction", hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction prediction for hash %s: %w", hash, err)
	}

	valueBig := baseRes[3].(*big.Int)
	feeBig := baseRes[4].(*big.Int)
	gasUsedBig := baseRes[5].(*big.Int)
	timestampBig := baseRes[6].(*big.Int)
	blockNumBig := baseRes[7].(*big.Int)

	valueF, _ := new(big.Float).SetInt(valueBig).Float64()
	feeF, _ := new(big.Float).SetInt(feeBig).Float64()

	isFraud := predRes[0].(bool)
	verdict := predRes[1].(string)
	flagReason := predRes[2].(string)
	riskScoreBig := predRes[3].(*big.Int)
	data := predRes[4].(string)
	status := predRes[5].(string)

	dTx := domain.Transaction{
		Hash:        baseRes[0].(string),
		FromAddress: baseRes[1].(string),
		ToAddress:   baseRes[2].(string),
		Value:       valueF,
		Fee:         feeF,
		GasUsed:     int(gasUsedBig.Uint64()),
		Timestamp:   time.Unix(timestampBig.Int64(), 0),
		BlockHeight: int(blockNumBig.Uint64()),
		IsFraud:     isFraud,
		Verdict:     verdict,
		FlagReason:  flagReason,
		RiskScore:   int(riskScoreBig.Uint64()),
		Data:        data,
		Status:      status,
	}

	return &dTx, nil
}

func (r *blockchainRepository) GetRecentTransactions(limit int) ([]domain.Transaction, error) {
	if limit <= 0 {
		limit = 10
	}

	countRes, err := r.callContractRead("getTransactionCount")
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction count: %w", err)
	}
	totalCount := countRes[0].(*big.Int).Int64()

	log.Printf("[Blockchain Repository] Retrieving %d recent transactions of total %d\n", limit, totalCount)

	var txs []domain.Transaction
	start := totalCount - 1
	end := totalCount - int64(limit)
	if end < 0 {
		end = 0
	}

	for i := start; i >= end; i-- {
		hashRes, err := r.callContractRead("getTransactionHashAtIndex", big.NewInt(i))
		if err != nil {
			continue
		}
		txHash := hashRes[0].(string)

		tx, err := r.GetTransactionByHash(txHash)
		if err == nil && tx != nil {
			txs = append(txs, *tx)
		}
	}

	return txs, nil
}

func (r *blockchainRepository) GetBalanceByAddress(address string) (float64, error) {
	account := common.HexToAddress(address)
	balance, err := r.client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return 0, err
	}

	fBalance := new(big.Float).SetInt(balance)
	ethValue := new(big.Float).Quo(fBalance, big.NewFloat(1e18))
	floatVal, _ := ethValue.Float64()
	return floatVal, nil
}

func (r *blockchainRepository) GetTransactionsByAddress(address string) ([]domain.Transaction, error) {
	countRes, err := r.callContractRead("getTransactionCount")
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction count: %w", err)
	}
	totalCount := countRes[0].(*big.Int).Int64()

	var txs []domain.Transaction
	addrHex := common.HexToAddress(address).Hex()

	log.Printf("[Blockchain Repository] Filtering transactions for address %s out of total %d\n", addrHex, totalCount)

	for i := int64(0); i < totalCount; i++ {
		hashRes, err := r.callContractRead("getTransactionHashAtIndex", big.NewInt(i))
		if err != nil {
			continue
		}
		txHash := hashRes[0].(string)

		tx, err := r.GetTransactionByHash(txHash)
		if err == nil && tx != nil {
			if common.HexToAddress(tx.FromAddress).Hex() == addrHex || common.HexToAddress(tx.ToAddress).Hex() == addrHex {
				txs = append(txs, *tx)
			}
		}
	}

	return txs, nil
}

func (r *blockchainRepository) RecordTransaction(txHash string, fromAddr string, toAddr string, value float64, fee float64, gasUsed int, data string) (string, error) {
	valueBig := big.NewInt(int64(value))
	feeBig := big.NewInt(int64(fee))
	gasUsedBig := big.NewInt(int64(gasUsed))

	evmHash, err := r.callContractWrite(
		"recordTransaction",
		txHash,
		fromAddr,
		toAddr,
		valueBig,
		feeBig,
		gasUsedBig,
		data,
	)
	if err != nil {
		return "", fmt.Errorf("failed to record transaction: %w", err)
	}

	return evmHash, nil
}

func (r *blockchainRepository) UpdateTransactionPrediction(txHash string, isFraud bool, riskScore int, verdict string, flagReason string) (string, error) {
	riskBig := big.NewInt(int64(riskScore))

	evmHash, err := r.callContractWrite(
		"updateTransactionPrediction",
		txHash,
		isFraud,
		riskBig,
		verdict,
		flagReason,
	)
	if err != nil {
		return "", fmt.Errorf("failed to update prediction: %w", err)
	}

	return evmHash, nil
}

func (r *blockchainRepository) GetBlockHeaderOnly(hashOrHeight string) (*domain.Block, error) {
	var block *types.Block
	var err error

	if height, errHeight := strconv.Atoi(hashOrHeight); errHeight == nil {
		block, err = r.client.BlockByNumber(context.Background(), big.NewInt(int64(height)))
	} else {
		block, err = r.client.BlockByHash(context.Background(), common.HexToHash(hashOrHeight))
	}

	if err != nil {
		return nil, err
	}

	return &domain.Block{
		Height:           int(block.NumberU64()),
		Hash:             block.Hash().Hex(),
		ParentHash:       block.ParentHash().Hex(),
		Timestamp:        time.Unix(int64(block.Time()), 0),
		Size:             int(block.Size()),
		Miner:            block.Coinbase().Hex(),
		TransactionCount: 0,
		Transactions:     nil,
	}, nil
}

func (r *blockchainRepository) GetStats() (int, int, error) {
	statsRes, err := r.callContractRead("getStats")
	if err != nil {
		return 0, 0, err
	}
	
	totalTxs := int(statsRes[0].(*big.Int).Int64())
	totalFraud := int(statsRes[1].(*big.Int).Int64())
	return totalTxs, totalFraud, nil
}

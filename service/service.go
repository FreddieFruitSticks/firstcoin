package service

import (
	"blockchain/coin"
	"blockchain/repository"
	"blockchain/utils"
	"blockchain/wallet"
	"fmt"
)

const (
	SeedDifficultyLevel = 5
)

type BlockchainService struct {
	Blockchain *coin.Blockchain
	Wallet     *wallet.Wallet
}

func NewBlockchainService(b *coin.Blockchain, w *wallet.Wallet) BlockchainService {
	return BlockchainService{
		Blockchain: b,
		Wallet:     w,
	}
}

func (s *BlockchainService) CreateNextBlock() (*coin.Block, *coin.Blockchain, error) {
	// coinbase transaction is the first transaction included by the miner
	coinbaseTransaction, _ := wallet.CreateCoinbaseTransaction(s.Wallet.Crypt)
	transactionPool := make([]repository.Transaction, 0)
	transactionPool = append(transactionPool, coinbaseTransaction)
	transactionPool = append(transactionPool, repository.GetTxPoolArray()...)
	block := s.Blockchain.GenerateNextBlock(&transactionPool)

	err := block.IsValidBlock(s.Blockchain.GetLastBlock())
	if err != nil {
		utils.ErrorLogger.Println(fmt.Sprintf("Error in createNextBlock. err: %s", err))
		return nil, nil, err
	}

	return &block, s.Blockchain, err
}

func (s *BlockchainService) CreateTx(receiverAddress []byte, amount int) (*repository.Transaction, error) {
	tx, _, err := s.Wallet.CreateTransaction(receiverAddress, amount)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// Need to validate the entire pool for two reasons:
// 1. Multiple individual transactions can be valid, while the entire pool is invalid. Eg, same spender spends all his money twice.
// 2. When a new block is mined, and added to the chain, and uTxOs are updated, current txs might no longer be valid (e.g. the UtxO was used).

//Note: Say there is a pair of txs that are invalid together, this will register the SECOND tx as the invalid one and keep the first.
func ValidateTxPoolDryRun(newTx *repository.Transaction) ([][]byte, error) {
	invalidTxIDs := make([][]byte, 0)
	var err error

	txPoolArray := repository.GetTxPoolArray()

	if newTx != nil {
		txPoolArray = append(txPoolArray, *newTx)
	}

	uTxOSetCopy := repository.CopyUTxOSet()

	for _, tx := range txPoolArray {
		if err = wallet.IsValidTransactionCopy(tx, uTxOSetCopy); err != nil {
			utils.ErrorLogger.Printf("error when validating txPool: %s\n", err)
			invalidTxIDs = append(invalidTxIDs, tx.ID)
			continue
		}

		for _, txIn := range tx.TxIns {
			repository.RemoveTxOFromUTxOCopy(repository.TxIDType(tx.ID), txIn, uTxOSetCopy)
		}
		repository.AddTxToUTxOSetCopy(tx, uTxOSetCopy)
	}

	return invalidTxIDs, err
}

// commit all block txs. Remove txs from current tx pool that exist in the block.
// Then further validate if the rest of the entire tx pool is valid and remove the txs that are invalid.
func CommitBlockTransactions(block coin.Block) error {
	if err := wallet.AreValidTransactions(block.Transactions); err != nil {
		return err
	}
	for _, tx := range block.Transactions {
		for _, txIn := range tx.TxIns {
			repository.RemoveTxOFromUTxOSet(repository.TxIDType(tx.ID), txIn)
		}
		repository.AddTxToUTxOSet(tx)
		repository.RemoveTxFromTxPool(tx.ID)
	}

	if invalidTxIDs, err := ValidateTxPoolDryRun(nil); err != nil {
		utils.InfoLogger.Println("Left over Tx pool is invalid after committing block. Emptying tx pool")
		for _, invalidTxID := range invalidTxIDs {
			repository.RemoveTxFromTxPool(invalidTxID)
		}
	}

	return nil
}

func CreateGenesisBlockchain(crypt wallet.Cryptographic, blockchain coin.Blockchain) (coin.Blockchain, repository.Transaction) {
	genesisTransactionPool := make([]repository.Transaction, 0)

	// coinbase transaction is the first transaction included by the miner
	coinbaseTransaction, _ := wallet.CreateCoinbaseTransaction(crypt)
	genesisTransactionPool = append(genesisTransactionPool, coinbaseTransaction)

	repository.AddTxToUTxOSet(coinbaseTransaction)

	blockchain.AddBlock(coin.GenesisBlock(SeedDifficultyLevel, genesisTransactionPool))

	return blockchain, coinbaseTransaction
}

func (s *BlockchainService) AddTxToTxPool(tx repository.Transaction) bool {

	repository.AddTxToTxPool(tx)

	return true
}

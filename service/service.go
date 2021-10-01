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
	coinbaseTransaction, _ := wallet.CreateCoinbaseTransaction(s.Wallet.Crypt, s.Blockchain.GetLastBlock().Index+1)
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
// 2. When a new block is mined, and added to the chain, and uTxOs are updated, current txs might no longer be valid.
func ValidateTxPoolDryRun(blockIndex int, newTx repository.Transaction) error {
	txPoolArray := repository.GetTxPoolArray()
	txPoolArray = append(txPoolArray, newTx)

	uTxOSetCopy := repository.CopyUTxOSet()

	for _, tx := range txPoolArray {
		if err := wallet.IsValidTransactionCopy(tx, uTxOSetCopy); err != nil {
			utils.ErrorLogger.Printf("error when validating txPool: %s\n", err)
			return err
		}

		for _, txIn := range tx.TxIns {
			repository.RemoveUTxOFromSenderCopy(txIn, uTxOSetCopy)
		}
		for _, txO := range tx.TxOuts {
			repository.AddTxOToReceiverCopy(tx.ID, blockIndex, txO, uTxOSetCopy)
		}
	}
	return nil
}

func CommitBlockTransactions(block coin.Block) {
	for _, tx := range block.Transactions {
		for _, txIn := range tx.TxIns {
			repository.RemoveUTxOFromSender(txIn)
		}
		for _, txO := range tx.TxOuts {
			repository.AddTxOToReceiver(tx.ID, block.Index, txO)
		}
	}

	// TODO: Just for now - need to only remove the txs that are in the block
	repository.EmptyTxPool()
}

func CreateGenesisBlockchain(crypt wallet.Cryptographic, blockchain coin.Blockchain) (coin.Blockchain, repository.Transaction) {
	genesisTransactionPool := make([]repository.Transaction, 0)

	// coinbase transaction is the first transaction included by the miner
	coinbaseTransaction, _ := wallet.CreateCoinbaseTransaction(crypt, 0)
	genesisTransactionPool = append(genesisTransactionPool, coinbaseTransaction)

	repository.AddTxOToReceiver(coinbaseTransaction.ID, 0, coinbaseTransaction.TxOuts[0])

	blockchain.AddBlock(coin.GenesisBlock(SeedDifficultyLevel, genesisTransactionPool))

	return blockchain, coinbaseTransaction
}

func (s *BlockchainService) AddTxToTxPool(tx repository.Transaction) bool {

	repository.AddTxToTxPool(tx)

	return true
}

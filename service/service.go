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
	transactionPool = append(transactionPool, repository.GetTxPool()...)
	block := s.Blockchain.GenerateNextBlock(&transactionPool)

	err := block.IsValidBlock(s.Blockchain.GetLastBlock())
	if err != nil {
		utils.ErrorLogger.Println(fmt.Sprintf("Error in createNextBlock. err: %s", err))
		return nil, nil, err
	}

	return &block, s.Blockchain, err
}

func (s *BlockchainService) SpendMoney(receiverAddress []byte, amount int) (*repository.Transaction, error) {
	transaction, _, err := s.Wallet.CreateTransaction(receiverAddress, amount)
	if err != nil {
		return nil, err
	}

	repository.AddTxToTxPool(*transaction)

	return transaction, nil
}

func CommitBlockTransactions(block coin.Block) {
	// At this point assume each transaction is valid - checked previously

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

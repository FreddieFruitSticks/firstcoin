package service

import (
	"blockchain/coin"
	"blockchain/wallet"
)

type BlockchainService struct {
	Account                    *wallet.Account
	Blockchain                 *coin.Blockchain
	UnconfirmedTransactionPool *[]wallet.Transaction
	UTxOs                      *map[string]map[string]wallet.UTxOut
}

func NewBlockchainService(a *wallet.Account, b *coin.Blockchain, u *[]wallet.Transaction, uTxO *map[string]map[string]wallet.UTxOut) BlockchainService {
	return BlockchainService{
		Account:                    a,
		Blockchain:                 b,
		UnconfirmedTransactionPool: u,
		UTxOs:                      uTxO,
	}
}

func (s *BlockchainService) CreateNextBlock() (bool, *coin.Block, *coin.Blockchain, *map[string]map[string]wallet.UTxOut) {
	// coinbase transaction is the first transaction included by the miner
	coinbaseTransaction := wallet.CreateCoinbaseTransaction(*s.Account, s.Blockchain.GetLastBlock().Index+1)
	transactionPool := make([]wallet.Transaction, 0)
	transactionPool = append(transactionPool, coinbaseTransaction)
	transactionPool = append(transactionPool, *s.UnconfirmedTransactionPool...)

	block := s.Blockchain.GenerateNextBlock(&transactionPool)

	valid := block.IsValidBlock(s.Blockchain.GetLastBlock())
	if !valid {
		return false, nil, nil, nil
	}

	s.Blockchain.AddBlock(block)

	s.UpdateUnspentTxOutputs(block)

	return true, &block, s.Blockchain, s.UTxOs
}

func (s *BlockchainService) UpdateUnspentTxOutputs(block coin.Block) {
	// At this point assume each transaction is valid - checked previously

	s.UpdateUTxOWithCoinbaseTransaction(block)

	for _, _ = range (block.Transactions)[1:] {
	}
	// in here loop through transactions and update unspentTxOuts
}

func (s *BlockchainService) UpdateUTxOWithCoinbaseTransaction(block coin.Block) bool {
	coinbaseTx := (block.Transactions)[0]

	ownerUTxOs := (*s.UTxOs)[string(coinbaseTx.TxOuts[0].Address)]
	uTxO := wallet.UTxOut{
		ID:      coinbaseTx.ID,
		Index:   block.Index,
		Address: coinbaseTx.TxOuts[0].Address,
		Amount:  coinbaseTx.TxOuts[0].Amount,
	}

	if ownerUTxOs == nil {
		txIDMap := make(map[string]wallet.UTxOut)
		txIDMap[string(coinbaseTx.ID)] = uTxO
		ownerUTxOs = txIDMap
	} else {
		ownerUTxOs[string(coinbaseTx.ID)] = uTxO
	}

	(*s.UTxOs)[string(coinbaseTx.TxOuts[0].Address)] = ownerUTxOs

	return true
}

func (s *BlockchainService) AddBlockToBlockchain(block coin.Block) bool {
	if block.IsValidBlock(s.Blockchain.GetLastBlock()) && block.ValidTimestampToNow() {
		s.Blockchain.AddBlock(block)
		s.UpdateUnspentTxOutputs(block)
		// TODO: when this node receives a valid block, it must remove transactions from its own pool that exist in the blocks transactions data
		return true
	}
	return false
}

func (s *BlockchainService) CreateTransaction(address []byte, amount int) wallet.Transaction {
	unspentTransaction := wallet.UTxOut{
		Address: []byte("1235asasdasda"),
		Amount:  1000,
	}

	transaction := wallet.CreateTransaction(address, amount, &unspentTransaction, s.Account)
	*s.UnconfirmedTransactionPool = append(*s.UnconfirmedTransactionPool, transaction)

	return transaction
}

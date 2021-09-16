package service

import (
	"blockchain/coin"
	"blockchain/wallet"
)

type BlockchainService struct {
	Account                    *wallet.Account
	Blockchain                 *coin.Blockchain
	UnconfirmedTransactionPool *[]wallet.Transaction
	UTxOSet                    *wallet.UTxOSetType
}

func NewBlockchainService(a *wallet.Account, b *coin.Blockchain, u *[]wallet.Transaction, uTxO *wallet.UTxOSetType) BlockchainService {
	return BlockchainService{
		Account:                    a,
		Blockchain:                 b,
		UnconfirmedTransactionPool: u,
		UTxOSet:                    uTxO,
	}
}

func (s *BlockchainService) CreateNextBlock() (bool, *coin.Block, *coin.Blockchain, *wallet.UTxOSetType) {
	// coinbase transaction is the first transaction included by the miner
	coinbaseTransaction, _ := wallet.CreateCoinbaseTransaction(*s.Account, s.Blockchain.GetLastBlock().Index+1)
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

	return true, &block, s.Blockchain, s.UTxOSet
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

	receiverUTxOs := (*s.UTxOSet)[wallet.PublicKeyAddressType(coinbaseTx.TxOuts[0].Address)]
	uTxO := wallet.UTxO{
		ID:      coinbaseTx.ID,
		Index:   block.Index,
		Address: coinbaseTx.TxOuts[0].Address,
		Amount:  coinbaseTx.TxOuts[0].Amount,
	}

	if receiverUTxOs == nil {
		txIDMap := make(map[wallet.TxIDType]wallet.UTxO)
		txIDMap[wallet.TxIDType(coinbaseTx.ID)] = uTxO
		receiverUTxOs = txIDMap
	} else {
		receiverUTxOs[wallet.TxIDType(coinbaseTx.ID)] = uTxO
	}

	(*s.UTxOSet)[wallet.PublicKeyAddressType(coinbaseTx.TxOuts[0].Address)] = receiverUTxOs

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
	unspentTransaction := wallet.UTxO{
		Address: []byte("1235asasdasda"),
		Amount:  1000,
	}

	transaction := wallet.CreateTransaction(address, amount, &unspentTransaction, s.Account)
	*s.UnconfirmedTransactionPool = append(*s.UnconfirmedTransactionPool, transaction)

	return transaction
}

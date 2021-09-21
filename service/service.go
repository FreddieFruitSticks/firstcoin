package service

import (
	"blockchain/coin"
	"blockchain/wallet"
)

type BlockchainService struct {
	Crypt                      *wallet.Cryptographic
	Blockchain                 *coin.Blockchain
	UnconfirmedTransactionPool *[]wallet.Transaction
	UTxOSet                    *wallet.UTxOSetType
	Wallet                     *wallet.Wallet
}

func NewBlockchainService(c *wallet.Cryptographic, b *coin.Blockchain, u *[]wallet.Transaction, uTxO *wallet.UTxOSetType, w *wallet.Wallet) BlockchainService {
	return BlockchainService{
		Crypt:                      c,
		Blockchain:                 b,
		UnconfirmedTransactionPool: u,
		UTxOSet:                    uTxO,
		Wallet:                     w,
	}
}

func (s *BlockchainService) CreateNextBlock() (bool, *coin.Block, *coin.Blockchain, *wallet.UTxOSetType) {
	// coinbase transaction is the first transaction included by the miner
	coinbaseTransaction, _ := wallet.CreateCoinbaseTransaction(*s.Crypt, s.Blockchain.GetLastBlock().Index+1)
	transactionPool := make([]wallet.Transaction, 0)
	transactionPool = append(transactionPool, coinbaseTransaction)
	transactionPool = append(transactionPool, *s.UnconfirmedTransactionPool...)

	block := s.Blockchain.GenerateNextBlock(&transactionPool)

	valid := block.IsValidBlock(s.Blockchain.GetLastBlock(), s.Wallet.UTxOSet)
	if !valid {
		return false, nil, nil, nil
	}

	s.Blockchain.AddBlock(block)

	s.UpdateUTxOSet(block)

	return true, &block, s.Blockchain, s.UTxOSet
}

func (s *BlockchainService) UpdateUTxOWithCoinbaseTransaction(block coin.Block) bool {
	coinbaseTx := (block.Transactions)[0]

	receiverUTxOs := (*s.UTxOSet)[wallet.PublicKeyAddressType(coinbaseTx.TxOuts[0].Address)]
	uTxO := wallet.UTxO{
		ID: wallet.UTxOID{
			TxID:    coinbaseTx.ID,
			Address: coinbaseTx.TxOuts[0].Address,
		},
		Index:  block.Index,
		Amount: coinbaseTx.TxOuts[0].Amount,
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

func (s *BlockchainService) ValidateAndAddBlockToBlockchain(block coin.Block) bool {
	if block.IsValidBlock(s.Blockchain.GetLastBlock(), s.Wallet.UTxOSet) && block.ValidTimestampToNow() {
		s.Blockchain.AddBlock(block)
		s.UpdateUTxOSet(block)
		// TODO: when this node receives a valid block, it must remove transactions from its own pool that exist in the blocks transactions data
		return true
	}
	return false
}

func (s *BlockchainService) SpendMoney(receiverAddress []byte, amount int) (*wallet.Transaction, error) {
	transaction, _, err := s.Wallet.CreateTransaction(receiverAddress, amount)
	if err != nil {
		return nil, err
	}

	*s.UnconfirmedTransactionPool = append(*s.UnconfirmedTransactionPool, *transaction)

	return transaction, nil
}

func (s *BlockchainService) UpdateUTxOSet(block coin.Block) {
	// At this point assume each transaction is valid - checked previously

	s.UpdateUTxOWithCoinbaseTransaction(block)

	for _, _ = range (block.Transactions)[1:] {
	}
	// in here loop through transactions and update unspentTxOuts
}

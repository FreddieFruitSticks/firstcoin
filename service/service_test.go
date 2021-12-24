package service_test

import (
	"blockchain/coin"
	"blockchain/repository"
	"blockchain/service"
	"blockchain/wallet"
	"testing"
)

func TestUpdateUTxOSet(t *testing.T) {
	t.Run("update legitimate tx", func(t *testing.T) {
		amount := 5

		senderCrypt := wallet.NewCryptographic()
		senderCrypt.GenerateKeyPair()

		receiverCrypt := wallet.NewCryptographic()
		receiverCrypt.GenerateKeyPair()

		senderWallet := wallet.NewWallet(*senderCrypt)

		blocks := make([]coin.Block, 0)
		blockchain := coin.NewBlockchain(blocks)
		*blockchain, _ = service.CreateGenesisBlockchain(*senderCrypt, *blockchain)

		coinbaseTransaction, _ := wallet.CreateCoinbaseTransaction(*senderCrypt)

		tx, _, err := senderWallet.CreateTransaction(receiverCrypt.PublicKey, amount)
		if err != nil {
			t.Fatalf(err.Error())
		}

		transactionPool := make([]repository.Transaction, 0)
		transactionPool = append(transactionPool, coinbaseTransaction)
		transactionPool = append(transactionPool, *tx)

		block := blockchain.GenerateNextBlock(&transactionPool)
		err = service.CommitBlockTransactions(block)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		uTxOSet := repository.GetEntireUTxOSet()
		if len(uTxOSet) != 2 {
			t.Fatalf("Length of uTxOSet incorrect. Got: %d. Want:%d", len(uTxOSet), 2)
		}

		txOutsOfReceived := uTxOSet[repository.TxIDType(tx.ID)].TxOuts
		if len(txOutsOfReceived) != 2 {
			t.Fatalf("Length of uTxOuts incorrect. Got: %d. Want:%d", len(txOutsOfReceived), 2)
		}

		if txOutsOfReceived[0].Value != 5 || txOutsOfReceived[1].Value != 5 {
			t.Fatalf("Length of uTxOuts incorrect. Got: %d. Want:%d", len(txOutsOfReceived), 2)
		}
	})
}

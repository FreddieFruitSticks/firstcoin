package service_test

import (
	"firstcoin/coin"
	"firstcoin/repository"
	"firstcoin/service"
	"firstcoin/wallet"
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

		coinbaseTransaction, _ := wallet.CreateCoinbaseTransaction(*senderCrypt, 1)

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

		paidAmount := txOutsOfReceived[0].Value
		change := txOutsOfReceived[1].Value

		if paidAmount != amount || change != wallet.COINBASE_TRANSACTION_AMOUNT-amount-wallet.TRANSACTION_FEE {
			t.Fatalf("txO received incorrect. Got: %d. Want:%d", txOutsOfReceived[0].Value, wallet.COINBASE_TRANSACTION_AMOUNT-amount-wallet.TRANSACTION_FEE)
		}
	})
}

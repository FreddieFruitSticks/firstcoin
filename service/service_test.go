package service_test

import (
	"blockchain/coin"
	"blockchain/repository"
	"blockchain/service"
	"blockchain/wallet"
	"fmt"
	"reflect"
	"testing"
)

func TestUpdateUTxOSet(t *testing.T) {
	t.Run("update legitimate tx", func(t *testing.T) {
		amount := 5

		senderCrypt := wallet.NewCryptographic()
		senderCrypt.GenerateKeyPair()

		receiverCrypt := wallet.NewCryptographic()
		receiverCrypt.GenerateKeyPair()

		expectedUTxOSet := repository.UTxOSetType(make(map[repository.PublicKeyAddressType]map[repository.TxIDType]repository.UTxO, 0))

		senderWallet := wallet.NewWallet(*senderCrypt)

		blocks := make([]coin.Block, 0)
		blockchain := coin.NewBlockchain(blocks)
		*blockchain, _ = service.CreateGenesisBlockchain(*senderCrypt, *blockchain)

		tx, _, err := senderWallet.CreateTransaction(receiverCrypt.PublicKey, amount)
		if err != nil {
			t.Fatalf(err.Error())
		}

		transactionPool := make([]wallet.Transaction, 0)
		transactionPool = append(transactionPool, *tx)

		block := blockchain.GenerateNextBlock(&transactionPool)
		service.CommitBlockTransactions(block)

		uTxO1 := repository.UTxO{
			ID: repository.UTxOID{
				Address: senderCrypt.PublicKey,
				TxID:    tx.ID,
			},
			Index:  1,
			Amount: 5,
		}

		uTxO2 := repository.UTxO{
			ID: repository.UTxOID{
				Address: receiverCrypt.PublicKey,
				TxID:    tx.ID,
			},
			Index:  1,
			Amount: 5,
		}

		uTxOTxIDMapSender := make(map[repository.TxIDType]repository.UTxO)
		uTxOTxIDMapSender[repository.TxIDType(tx.ID)] = uTxO1
		uTxOTxIDMapReceiver := make(map[repository.TxIDType]repository.UTxO)
		uTxOTxIDMapReceiver[repository.TxIDType(tx.ID)] = uTxO2

		expectedUTxOSet[repository.PublicKeyAddressType(senderCrypt.PublicKey)] = uTxOTxIDMapSender
		expectedUTxOSet[repository.PublicKeyAddressType(receiverCrypt.PublicKey)] = uTxOTxIDMapReceiver

		if !reflect.DeepEqual(expectedUTxOSet, repository.GetEntireUTxOSet()) {
			fmt.Printf("%+v\n", repository.GetEntireUTxOSet())
			fmt.Printf("-------------------\n")
			fmt.Printf("%+v\n", expectedUTxOSet)
			t.Fatalf("")
		}

	})
}

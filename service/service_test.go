package service_test

import (
	"blockchain/coin"
	"blockchain/repository"
	"blockchain/service"
	"blockchain/wallet"
	"fmt"
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

		uTxO1 := repository.UTxO{
			ID:     repository.UTxOID{},
			Index:  0,
			Amount: 10,
		}

		// uTxO2 := repository.UTxO{
		// 	ID: repository.UTxOID{
		// 		TxID:    previousTxID2,
		// 		Address: senderCrypt.PublicKey,
		// 	},
		// 	Index:  1,
		// 	Amount: 50,
		// }

		uTxOTxIDMap := make(map[repository.TxIDType]repository.UTxO)
		uTxOTxIDMap[repository.TxIDType(senderCrypt.PublicKey)] = uTxO1
		expectedUTxOSet[repository.PublicKeyAddressType(senderCrypt.PublicKey)] = uTxOTxIDMap

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

		// bs := service.NewBlockchainService(blockchain, &transactionPool, &uTxOSet, senderWallet)

		// _ = blockchain.GenerateNextBlock(&transactionPool)
		block := blockchain.GenerateNextBlock(&transactionPool)
		service.CommitBlockTransactions(block)

		// if !reflect.DeepEqual(repository.GetEntireUTxOSet(), expectedUTxOSet) {
		// 	fmt.Printf("%+v\n", repository.GetEntireUTxOSet())
		// 	fmt.Printf("-------------------\n")
		// 	fmt.Printf("%+v\n", expectedUTxOSet)
		// 	t.Fatalf("")
		// }

		senderLedger := repository.GetUserLedger(senderCrypt.PublicKey)
		receiverLedger := repository.GetUserLedger(receiverCrypt.PublicKey)

		fmt.Printf("%+v\n", senderLedger)
		fmt.Println("----------------")
		fmt.Printf("%+v\n", receiverLedger)

		// expectedUTxO1 := repository.UTxO{
		// 	ID: repository.UTxOID{
		// 		TxID:    previousTxID,
		// 		Address: senderCrypt.PublicKey,
		// 	},
		// 	Index:  1,
		// 	Amount: 70,
		// }

		// uTxO2 := repository.UTxO{
		// 	ID: repository.UTxOID{
		// 		TxID:    previousTxID2,
		// 		Address: senderCrypt.PublicKey,
		// 	},
		// 	Index:  1,
		// 	Amount: 50,
		// }

		// uTxOTxIDMap := make(map[repository.TxIDType]repository.UTxO)
		// uTxOTxIDMap[repository.TxIDType(previousTxID)] = uTxO1
		// uTxOTxIDMap[repository.TxIDType(previousTxID2)] = uTxO2
		// uTxOSet[repository.PublicKeyAddressType(senderCrypt.PublicKey)] = uTxOTxIDMap

	})
}

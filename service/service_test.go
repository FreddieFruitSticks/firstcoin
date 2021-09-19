package service_test

import (
	"blockchain/coin"
	"blockchain/service"
	"blockchain/wallet"
	"reflect"
	"testing"
)

func TestFindUTxOs(test *testing.T) {
	crypt := wallet.NewCryptographic()
	crypt.GenerateKeyPair()
	blocks := make([]coin.Block, 0)
	blockchain := coin.NewBlockchain(blocks)
	transactionPool := make([]wallet.Transaction, 0)

	test.Run("UTxOs can service the amount", func(t *testing.T) {
		uTxOSet := wallet.UTxOSetType(make(map[wallet.PublicKeyAddressType]map[wallet.TxIDType]wallet.UTxO, 0))
		service := service.NewBlockchainService(crypt, blockchain, &transactionPool, &uTxOSet)

		id1 := []byte{1, 2, 3}
		uTxO1 := wallet.UTxO{
			ID: wallet.UTxOID{
				Address: crypt.PublicKey,
				TxID:    id1,
			},
			Amount: 70,
			Index:  1,
		}

		id2 := []byte{4, 5, 6}
		uTxO2 := wallet.UTxO{
			ID: wallet.UTxOID{
				Address: crypt.PublicKey,
				TxID:    id1,
			},
			Amount: 50,
			Index:  1,
		}

		uTxOMap := make(map[wallet.TxIDType]wallet.UTxO)
		uTxOMap[wallet.TxIDType(id1)] = uTxO1
		uTxOMap[wallet.TxIDType(id2)] = uTxO2

		uTxOSet[wallet.PublicKeyAddressType(crypt.PublicKey)] = uTxOMap

		expectedUtxOs := make([]wallet.UTxO, 0)
		expectedUtxOs = append(expectedUtxOs, uTxO1)
		expectedUtxOs = append(expectedUtxOs, uTxO2)

		// expectedUtxOs := []wallet.UTxO{uTxO1, uTxO2}

		uTxOs, err := service.FindUTxOs(100)
		if err != nil {
			t.Fatalf("expected to find UTxOs. err: %s", err)
		}

		if !reflect.DeepEqual(uTxOs, expectedUtxOs) {
			t.Fatalf("incorrect UtxOs\nGot:%+v\nWant:%+v", uTxOs, expectedUtxOs)
		}
	})

	test.Run("UTxOs cannot service the amount - insufficient funds", func(t *testing.T) {
		uTxOSet := wallet.UTxOSetType(make(map[wallet.PublicKeyAddressType]map[wallet.TxIDType]wallet.UTxO, 0))
		service := service.NewBlockchainService(crypt, blockchain, &transactionPool, &uTxOSet)

		id1 := []byte{1, 2, 3}
		uTxO1 := wallet.UTxO{
			ID: wallet.UTxOID{
				Address: crypt.PublicKey,
				TxID:    id1,
			},
			Amount: 70,
			Index:  1,
		}

		id2 := []byte{4, 5, 6}
		uTxO2 := wallet.UTxO{
			ID: wallet.UTxOID{
				Address: crypt.PublicKey,
				TxID:    id1,
			},
			Amount: 20,
			Index:  1,
		}

		uTxOMap := make(map[wallet.TxIDType]wallet.UTxO)
		uTxOMap[wallet.TxIDType(id1)] = uTxO1
		uTxOMap[wallet.TxIDType(id2)] = uTxO2

		uTxOSet[wallet.PublicKeyAddressType(crypt.PublicKey)] = uTxOMap

		_, err := service.FindUTxOs(100)
		if err == nil {
			t.Fatalf("expected to find err")
		}

		if err != nil && err.Error() != "insufficient funds" {
			t.Fatalf("incorrect error type, expected insufficient funds, got: %+v", err)
		}

	})
}

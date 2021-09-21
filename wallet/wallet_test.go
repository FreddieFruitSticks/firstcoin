package wallet_test

import (
	"blockchain/wallet"
	"reflect"
	"testing"
)

func TestCreateCoinbaseTransaction(t *testing.T) {
	t.Run("validate successful coinbase transaction", func(t *testing.T) {
		crypt := wallet.NewCryptographic()
		crypt.GenerateKeyPair()

		coinbaseTx, now := wallet.CreateCoinbaseTransaction(*crypt, 1)

		txIns := make([]wallet.TxIn, 0)
		txOuts := make([]wallet.TxO, 0)
		txIn := wallet.TxIn{
			UTxOID: wallet.UTxOID{
				Address: []byte{},
				TxID:    []byte{},
			},
			UTxOIndex: 1,
			Signature: []byte{},
		}

		txIns = append(txIns, txIn)

		txOut := wallet.TxO{
			Amount:  wallet.COINBASE_TRANSACTION_AMOUNT,
			Address: crypt.PublicKey,
		}
		txOuts = append(txOuts, txOut)

		expectedTx := wallet.Transaction{
			ID:        []byte{},
			TxIns:     txIns,
			TxOuts:    txOuts,
			Timestamp: now,
		}

		txID := wallet.GenerateTransactionID(expectedTx)
		expectedTx.ID = txID

		txInSignature := crypt.GenerateSignature(txID)
		txIns[0].Signature = txInSignature

		if !wallet.IsValidCoinbaseTransaction(coinbaseTx, 1) {
			t.Fatalf("coinbase Tx not valid")
		}

		if !reflect.DeepEqual(expectedTx.Timestamp, coinbaseTx.Timestamp) {
			t.Fatalf("coinbase Tx timestamp not equal to expected Tx timestamp")
		}

		if !reflect.DeepEqual(expectedTx.TxOuts, coinbaseTx.TxOuts) {
			t.Fatalf("coinbase Tx TxOuts not equal to expected Tx TxOuts")
		}

		if !reflect.DeepEqual(expectedTx.TxIns[0].UTxOID, coinbaseTx.TxIns[0].UTxOID) {
			t.Fatalf("coinbase Tx Ins UTxOID not equal to expected Tx Ins UTxOID")
		}

		if !reflect.DeepEqual(expectedTx.TxIns[0].UTxOIndex, coinbaseTx.TxIns[0].UTxOIndex) {
			t.Fatalf("coinbase Tx Ins UTxOIndex not equal to expected Tx Ins UTxOIndex")
		}

		expectedTxSigErr := wallet.VerifySignature(expectedTx.TxIns[0].Signature, crypt.PublicKey, wallet.GenerateTransactionID(expectedTx))
		coinbaseTxSigErr := wallet.VerifySignature(coinbaseTx.TxIns[0].Signature, crypt.PublicKey, wallet.GenerateTransactionID(coinbaseTx))

		if expectedTxSigErr != nil || coinbaseTxSigErr != nil {
			t.Fatalf("coinbase Tx Signatures invalid")
		}
	})
}

func TestCreateTransaction(t *testing.T) {
	t.Run("validate successful transaction - 2 inputs with change", func(t *testing.T) {
		amount := 100
		// uTxOAmount := 200

		// the output of a previous transaction that was sent to the sender of this transaction.
		previousTxID := []byte{1, 2, 3}
		previousTxID2 := []byte{1, 2, 3, 4}

		senderCrypt := wallet.NewCryptographic()
		senderCrypt.GenerateKeyPair()

		receiverCrypt := wallet.NewCryptographic()
		receiverCrypt.GenerateKeyPair()

		uTxOSet := wallet.UTxOSetType(make(map[wallet.PublicKeyAddressType]map[wallet.TxIDType]wallet.UTxO, 0))

		txIns := make([]wallet.TxIn, 0)
		txOuts := make([]wallet.TxO, 0)

		txIn := wallet.TxIn{
			UTxOID: wallet.UTxOID{
				Address: senderCrypt.PublicKey,
				TxID:    previousTxID,
			},
			UTxOIndex: 1,
			Signature: []byte{},
		}

		txIn2 := wallet.TxIn{
			UTxOID: wallet.UTxOID{
				Address: senderCrypt.PublicKey,
				TxID:    previousTxID2,
			},
			UTxOIndex: 1,
			Signature: []byte{},
		}

		txIns = append(txIns, txIn)
		txIns = append(txIns, txIn2)

		txOutReceiver := wallet.TxO{
			Address: receiverCrypt.PublicKey,
			Amount:  amount,
		}
		txOutSenderChange := wallet.TxO{
			Address: senderCrypt.PublicKey,
			Amount:  20,
		}
		txOuts = append(txOuts, txOutReceiver)
		txOuts = append(txOuts, txOutSenderChange)

		expectedSenderTx := wallet.Transaction{
			ID:     []byte{},
			TxIns:  txIns,
			TxOuts: txOuts,
		}

		uTxO1 := wallet.UTxO{
			ID: wallet.UTxOID{
				TxID:    previousTxID,
				Address: senderCrypt.PublicKey,
			},
			Index:  1,
			Amount: 70,
		}

		uTxO2 := wallet.UTxO{
			ID: wallet.UTxOID{
				TxID:    previousTxID2,
				Address: senderCrypt.PublicKey,
			},
			Index:  1,
			Amount: 50,
		}

		uTxOTxIDMap := make(map[wallet.TxIDType]wallet.UTxO)
		uTxOTxIDMap[wallet.TxIDType(previousTxID)] = uTxO1
		uTxOTxIDMap[wallet.TxIDType(previousTxID2)] = uTxO2
		uTxOSet[wallet.PublicKeyAddressType(senderCrypt.PublicKey)] = uTxOTxIDMap

		userWallet := wallet.NewWallet(&uTxOSet, *senderCrypt)

		tx, now, _ := userWallet.CreateTransaction(receiverCrypt.PublicKey, amount)

		expectedSenderTx.Timestamp = now

		expectedTxID := wallet.GenerateTransactionID(expectedSenderTx)
		// uTxO.ID = expectedTxID

		// senderWallet := wallet.NewWallet(uTxOSet, *senderCrypt)

		// this should actually be the txID of a different, older, tx.
		expectedSenderTx.ID = expectedTxID

		txIns[0].UTxOID.TxID = previousTxID
		txIns[0].Signature = senderCrypt.GenerateSignature(expectedTxID)
		txIns[1].Signature = senderCrypt.GenerateSignature(expectedTxID)

		if err := wallet.IsValidTransaction(expectedSenderTx, &uTxOSet); err != nil {
			t.Fatalf("Test failed: %+v", err)
		}

		if !reflect.DeepEqual(expectedSenderTx.Timestamp, tx.Timestamp) {
			t.Fatalf("Tx timestamp not equal to expected Tx timestamp")
		}

		if !reflect.DeepEqual(len(expectedSenderTx.TxOuts), len(tx.TxOuts)) {
			t.Fatalf("Tx TxOuts not equal to expected Tx TxOuts")
		}

		if !reflect.DeepEqual(expectedSenderTx.TxIns[0].UTxOID, tx.TxIns[0].UTxOID) {
			t.Fatalf("Tx Ins UTxOID not equal to expected Tx Ins UTxOID")
		}

		if !reflect.DeepEqual(expectedSenderTx.TxIns[0].UTxOIndex, tx.TxIns[0].UTxOIndex) {
			t.Fatalf("Tx Ins UTxOIndex not equal to expected Tx Ins UTxOIndex")
		}
	})

	t.Run("invalidate bad tx - not enough money", func(t *testing.T) {
		amount := 100
		// uTxOAmount := 200

		// the output of a previous transaction that was sent to the sender of this transaction.
		previousTxID := []byte{1, 2, 3}
		previousTxID2 := []byte{1, 2, 3, 4}

		senderCrypt := wallet.NewCryptographic()
		senderCrypt.GenerateKeyPair()

		receiverCrypt := wallet.NewCryptographic()
		receiverCrypt.GenerateKeyPair()

		uTxOSet := wallet.UTxOSetType(make(map[wallet.PublicKeyAddressType]map[wallet.TxIDType]wallet.UTxO, 0))

		txIns := make([]wallet.TxIn, 0)
		txOuts := make([]wallet.TxO, 0)

		txIn := wallet.TxIn{
			UTxOID: wallet.UTxOID{
				Address: senderCrypt.PublicKey,
				TxID:    previousTxID,
			},
			UTxOIndex: 1,
			Signature: []byte{},
		}

		txIn2 := wallet.TxIn{
			UTxOID: wallet.UTxOID{
				Address: senderCrypt.PublicKey,
				TxID:    previousTxID2,
			},
			UTxOIndex: 1,
			Signature: []byte{},
		}

		txIns = append(txIns, txIn)
		txIns = append(txIns, txIn2)

		txOutReceiver := wallet.TxO{
			Address: receiverCrypt.PublicKey,
			Amount:  amount,
		}
		txOutSenderChange := wallet.TxO{
			Address: senderCrypt.PublicKey,
			Amount:  20,
		}
		txOuts = append(txOuts, txOutReceiver)
		txOuts = append(txOuts, txOutSenderChange)

		uTxO1 := wallet.UTxO{
			ID: wallet.UTxOID{
				TxID:    previousTxID,
				Address: senderCrypt.PublicKey,
			},
			Index:  1,
			Amount: 70,
		}

		uTxO2 := wallet.UTxO{
			ID: wallet.UTxOID{
				TxID:    previousTxID2,
				Address: senderCrypt.PublicKey,
			},
			Index:  1,
			Amount: 20,
		}

		uTxOTxIDMap := make(map[wallet.TxIDType]wallet.UTxO)
		uTxOTxIDMap[wallet.TxIDType(previousTxID)] = uTxO1
		uTxOTxIDMap[wallet.TxIDType(previousTxID2)] = uTxO2
		uTxOSet[wallet.PublicKeyAddressType(senderCrypt.PublicKey)] = uTxOTxIDMap

		userWallet := wallet.NewWallet(&uTxOSet, *senderCrypt)

		_, _, err := userWallet.CreateTransaction(receiverCrypt.PublicKey, amount)
		if err == nil {
			t.Fatalf("Test failed: expected insufficient funds error")
		}
	})

	t.Run("invalidate bad tx - signature incorrect", func(t *testing.T) {
		amount := 100
		// uTxOAmount := 200

		// the output of a previous transaction that was sent to the sender of this transaction.
		previousTxID := []byte{1, 2, 3}
		previousTxID2 := []byte{1, 2, 3, 4}

		senderCrypt := wallet.NewCryptographic()
		senderCrypt.GenerateKeyPair()

		receiverCrypt := wallet.NewCryptographic()
		receiverCrypt.GenerateKeyPair()

		uTxOSet := wallet.UTxOSetType(make(map[wallet.PublicKeyAddressType]map[wallet.TxIDType]wallet.UTxO, 0))

		txIns := make([]wallet.TxIn, 0)
		txOuts := make([]wallet.TxO, 0)

		txIn := wallet.TxIn{
			UTxOID: wallet.UTxOID{
				Address: senderCrypt.PublicKey,
				TxID:    previousTxID,
			},
			UTxOIndex: 1,
			Signature: []byte{},
		}

		txIn2 := wallet.TxIn{
			UTxOID: wallet.UTxOID{
				Address: senderCrypt.PublicKey,
				TxID:    previousTxID2,
			},
			UTxOIndex: 1,
			Signature: []byte{},
		}

		txIns = append(txIns, txIn)
		txIns = append(txIns, txIn2)

		txOutReceiver := wallet.TxO{
			Address: receiverCrypt.PublicKey,
			Amount:  amount,
		}
		txOutSenderChange := wallet.TxO{
			Address: senderCrypt.PublicKey,
			Amount:  20,
		}
		txOuts = append(txOuts, txOutReceiver)
		txOuts = append(txOuts, txOutSenderChange)

		expectedSenderTx := wallet.Transaction{
			ID:     []byte{},
			TxIns:  txIns,
			TxOuts: txOuts,
		}

		uTxO1 := wallet.UTxO{
			ID: wallet.UTxOID{
				TxID:    previousTxID,
				Address: senderCrypt.PublicKey,
			},
			Index:  1,
			Amount: 70,
		}

		uTxO2 := wallet.UTxO{
			ID: wallet.UTxOID{
				TxID:    previousTxID2,
				Address: senderCrypt.PublicKey,
			},
			Index:  1,
			Amount: 20,
		}

		uTxOTxIDMap := make(map[wallet.TxIDType]wallet.UTxO)
		uTxOTxIDMap[wallet.TxIDType(previousTxID)] = uTxO1
		uTxOTxIDMap[wallet.TxIDType(previousTxID2)] = uTxO2
		uTxOSet[wallet.PublicKeyAddressType(senderCrypt.PublicKey)] = uTxOTxIDMap

		userWallet := wallet.NewWallet(&uTxOSet, *senderCrypt)

		_, now, err := userWallet.CreateTransaction(receiverCrypt.PublicKey, amount)
		if err == nil {
			t.Fatalf("Test failed: expected insufficient funds error")
		}

		expectedSenderTx.Timestamp = now

		expectedTxID := wallet.GenerateTransactionID(expectedSenderTx)
		// uTxO.ID = expectedTxID

		// senderWallet := wallet.NewWallet(uTxOSet, *senderCrypt)

		// this should actually be the txID of a different, older, tx.
		expectedSenderTx.ID = expectedTxID

		txIns[0].UTxOID.TxID = previousTxID
		txIns[0].Signature = senderCrypt.GenerateSignature(expectedTxID)
		txIns[1].Signature = []byte{1, 2, 3}

		if err := wallet.IsValidTransaction(expectedSenderTx, &uTxOSet); err == nil ||
			(err != nil && err.Error() != "Invalid transaction - signature verification failed: crypto/rsa: verification error") {
			t.Fatalf("Test failed: expected signature check to fail.\nGot: %s", err.Error())
		}
	})
}

func TestFindUTxOs(test *testing.T) {
	crypt := wallet.NewCryptographic()
	crypt.GenerateKeyPair()

	test.Run("UTxOs can service the amount", func(t *testing.T) {
		uTxOSet := wallet.UTxOSetType(make(map[wallet.PublicKeyAddressType]map[wallet.TxIDType]wallet.UTxO, 0))
		userWallet := wallet.NewWallet(&uTxOSet, *crypt)

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
				TxID:    id2,
			},
			Amount: 50,
			Index:  1,
		}

		uTxOMap := make(map[wallet.TxIDType]wallet.UTxO)
		uTxOMap[wallet.TxIDType(id2)] = uTxO2
		uTxOMap[wallet.TxIDType(id1)] = uTxO1

		uTxOSet[wallet.PublicKeyAddressType(crypt.PublicKey)] = uTxOMap

		expectedUtxOs := make([]wallet.UTxO, 0)
		expectedUtxOs = append(expectedUtxOs, uTxO2)
		expectedUtxOs = append(expectedUtxOs, uTxO1)

		// expectedUtxOs := []wallet.UTxO{uTxO1, uTxO2}

		uTxOs, _, err := userWallet.FindUTxOs(100)
		if err != nil {
			t.Fatalf("expected to find UTxOs. err: %s", err)
		}

		if !reflect.DeepEqual(uTxOs, expectedUtxOs) {
			t.Fatalf("incorrect UtxOs\nGot:%+v\nWant:%+v", uTxOs, expectedUtxOs)
		}
	})

	test.Run("UTxOs cannot service the amount - insufficient funds", func(t *testing.T) {
		uTxOSet := wallet.UTxOSetType(make(map[wallet.PublicKeyAddressType]map[wallet.TxIDType]wallet.UTxO, 0))
		userWallet := wallet.NewWallet(&uTxOSet, *crypt)

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
				TxID:    id2,
			},
			Amount: 20,
			Index:  1,
		}

		uTxOMap := make(map[wallet.TxIDType]wallet.UTxO)
		uTxOMap[wallet.TxIDType(id1)] = uTxO1
		uTxOMap[wallet.TxIDType(id2)] = uTxO2

		uTxOSet[wallet.PublicKeyAddressType(crypt.PublicKey)] = uTxOMap

		_, _, err := userWallet.FindUTxOs(100)
		if err == nil {
			t.Fatalf("expected to find err")
		}

		if err != nil && err.Error() != "insufficient funds" {
			t.Fatalf("incorrect error type, expected insufficient funds, got: %+v", err)
		}

	})
}

func TestGetTxOs(test *testing.T) {
	crypt := wallet.NewCryptographic()
	crypt.GenerateKeyPair()

	test.Run("UTxOs can service the amount", func(t *testing.T) {
		uTxOSet := wallet.UTxOSetType(make(map[wallet.PublicKeyAddressType]map[wallet.TxIDType]wallet.UTxO, 0))
		userWallet := wallet.NewWallet(&uTxOSet, *crypt)

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
				TxID:    id2,
			},
			Amount: 50,
			Index:  1,
		}

		uTxOs := make([]wallet.UTxO, 0)
		uTxOs = append(uTxOs, uTxO1)
		uTxOs = append(uTxOs, uTxO2)

		receiverAddress := []byte{1, 2, 3}

		expectedTxOs := make([]wallet.TxO, 0)
		expectedTxOs = append(expectedTxOs, wallet.TxO{
			Address: receiverAddress,
			Amount:  100,
		})

		expectedTxOs = append(expectedTxOs, wallet.TxO{
			Address: crypt.PublicKey,
			Amount:  20,
		})

		txOs, err := userWallet.GetTxOs(100, receiverAddress, uTxOs)
		if err != nil {
			t.Fatalf("expected to find UTxOs. err: %s", err)
		}

		if !reflect.DeepEqual(txOs, expectedTxOs) {
			t.Fatalf("incorrect UtxOs\nGot:%+v\nWant:%+v", uTxOs, expectedTxOs)
		}
	})

	test.Run("UTxOs cannot service the amount - insufficient funds", func(t *testing.T) {
		uTxOSet := wallet.UTxOSetType(make(map[wallet.PublicKeyAddressType]map[wallet.TxIDType]wallet.UTxO, 0))
		userWallet := wallet.NewWallet(&uTxOSet, *crypt)

		id1 := []byte{1, 2, 3}
		uTxO1 := wallet.UTxO{
			ID: wallet.UTxOID{
				Address: crypt.PublicKey,
				TxID:    id1,
			},
			Amount: 20,
			Index:  1,
		}

		id2 := []byte{4, 5, 6}
		uTxO2 := wallet.UTxO{
			ID: wallet.UTxOID{
				Address: crypt.PublicKey,
				TxID:    id2,
			},
			Amount: 50,
			Index:  1,
		}

		uTxOs := make([]wallet.UTxO, 0)
		uTxOs = append(uTxOs, uTxO1)
		uTxOs = append(uTxOs, uTxO2)

		receiverAddress := []byte{1, 2, 3}

		expectedTxOs := make([]wallet.TxO, 0)
		expectedTxOs = append(expectedTxOs, wallet.TxO{
			Address: receiverAddress,
			Amount:  100,
		})

		expectedTxOs = append(expectedTxOs, wallet.TxO{
			Address: crypt.PublicKey,
			Amount:  20,
		})

		_, err := userWallet.GetTxOs(100, receiverAddress, uTxOs)
		if err == nil {
			t.Fatalf("expected error of insufficient funds")
		}

	})
}

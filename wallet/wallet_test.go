package wallet_test

import (
	"blockchain/repository"
	"blockchain/wallet"
	"reflect"
	"testing"
)

func TestCreateCoinbaseTransaction(t *testing.T) {
	t.Run("validate successful coinbase transaction", func(t *testing.T) {
		crypt := wallet.NewCryptographic()
		crypt.GenerateKeyPair()

		coinbaseTx, now := wallet.CreateCoinbaseTransaction(*crypt, 1)

		txIns := make([]repository.TxIn, 0)
		txOuts := make([]repository.TxO, 0)
		txIn := repository.TxIn{
			UTxOID: repository.UTxOID{
				Address: []byte{},
				TxID:    []byte{},
			},
			UTxOIndex: 1,
			Signature: []byte{},
		}

		txIns = append(txIns, txIn)

		txOut := repository.TxO{
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

		if err := wallet.IsValidCoinbaseTransaction(coinbaseTx, 1); err != nil {
			t.Fatalf("coinbase Tx not valid %s", err.Error())
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

		txIns := make([]repository.TxIn, 0)
		txOuts := make([]repository.TxO, 0)

		txIn := repository.TxIn{
			UTxOID: repository.UTxOID{
				Address: senderCrypt.PublicKey,
				TxID:    previousTxID,
			},
			UTxOIndex: 1,
			Signature: []byte{},
		}

		txIn2 := repository.TxIn{
			UTxOID: repository.UTxOID{
				Address: senderCrypt.PublicKey,
				TxID:    previousTxID2,
			},
			UTxOIndex: 1,
			Signature: []byte{},
		}

		txIns = append(txIns, txIn)
		txIns = append(txIns, txIn2)

		txOutReceiver := repository.TxO{
			Address: receiverCrypt.PublicKey,
			Amount:  amount,
		}
		txOutSenderChange := repository.TxO{
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

		txO1 := repository.TxO{
			Address: senderCrypt.PublicKey,
			Amount:  70,
		}

		txO2 := repository.TxO{
			Address: senderCrypt.PublicKey,
			Amount:  50,
		}

		repository.AddTxOToReceiver(previousTxID, 1, txO1)
		repository.AddTxOToReceiver(previousTxID2, 1, txO2)

		userWallet := wallet.NewWallet(*senderCrypt)

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

		if err := wallet.IsValidTransaction(expectedSenderTx); err != nil {
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

		txIns := make([]repository.TxIn, 0)
		txOuts := make([]repository.TxO, 0)

		txIn := repository.TxIn{
			UTxOID: repository.UTxOID{
				Address: senderCrypt.PublicKey,
				TxID:    previousTxID,
			},
			UTxOIndex: 1,
			Signature: []byte{},
		}

		txIn2 := repository.TxIn{
			UTxOID: repository.UTxOID{
				Address: senderCrypt.PublicKey,
				TxID:    previousTxID2,
			},
			UTxOIndex: 1,
			Signature: []byte{},
		}

		txIns = append(txIns, txIn)
		txIns = append(txIns, txIn2)

		txOutReceiver := repository.TxO{
			Address: receiverCrypt.PublicKey,
			Amount:  amount,
		}
		txOutSenderChange := repository.TxO{
			Address: senderCrypt.PublicKey,
			Amount:  20,
		}
		txOuts = append(txOuts, txOutReceiver)
		txOuts = append(txOuts, txOutSenderChange)

		txO1 := repository.TxO{
			Address: senderCrypt.PublicKey,
			Amount:  70,
		}

		txO2 := repository.TxO{
			Address: senderCrypt.PublicKey,
			Amount:  20,
		}

		repository.AddTxOToReceiver(previousTxID, 1, txO1)
		repository.AddTxOToReceiver(previousTxID2, 1, txO2)

		userWallet := wallet.NewWallet(*senderCrypt)

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

		txIns := make([]repository.TxIn, 0)
		txOuts := make([]repository.TxO, 0)

		txIn := repository.TxIn{
			UTxOID: repository.UTxOID{
				Address: senderCrypt.PublicKey,
				TxID:    previousTxID,
			},
			UTxOIndex: 1,
			Signature: []byte{},
		}

		txIn2 := repository.TxIn{
			UTxOID: repository.UTxOID{
				Address: senderCrypt.PublicKey,
				TxID:    previousTxID2,
			},
			UTxOIndex: 1,
			Signature: []byte{},
		}

		txIns = append(txIns, txIn)
		txIns = append(txIns, txIn2)

		txOutReceiver := repository.TxO{
			Address: receiverCrypt.PublicKey,
			Amount:  amount,
		}
		txOutSenderChange := repository.TxO{
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

		txO1 := repository.TxO{
			Address: senderCrypt.PublicKey,
			Amount:  70,
		}

		txO2 := repository.TxO{
			Address: senderCrypt.PublicKey,
			Amount:  20,
		}

		repository.AddTxOToReceiver(previousTxID, 1, txO1)
		repository.AddTxOToReceiver(previousTxID2, 1, txO2)

		userWallet := wallet.NewWallet(*senderCrypt)

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

		if err := wallet.IsValidTransaction(expectedSenderTx); err == nil ||
			(err != nil && err.Error() != "Invalid transaction - signature verification failed: crypto/rsa: verification error") {
			t.Fatalf("Test failed: expected signature check to fail.\nGot: %s", err.Error())
		}
	})
}

func TestFindUTxOs(test *testing.T) {
	crypt := wallet.NewCryptographic()
	crypt.GenerateKeyPair()
	test.Run("UTxOs can service the amount", func(t *testing.T) {
		crypt := wallet.NewCryptographic()
		crypt.GenerateKeyPair()
		userWallet := wallet.NewWallet(*crypt)

		id1 := []byte{1, 2, 3}
		uTxO1 := repository.UTxO{
			ID: repository.UTxOID{
				Address: crypt.PublicKey,
				TxID:    id1,
			},
			Amount: 70,
			Index:  1,
		}

		id2 := []byte{4, 5, 6}
		uTxO2 := repository.UTxO{
			ID: repository.UTxOID{
				Address: crypt.PublicKey,
				TxID:    id2,
			},
			Amount: 50,
			Index:  1,
		}

		txO1 := repository.TxO{
			Address: crypt.PublicKey,
			Amount:  70,
		}

		txO2 := repository.TxO{
			Address: crypt.PublicKey,
			Amount:  50,
		}

		repository.AddTxOToReceiver(id1, 1, txO1)
		repository.AddTxOToReceiver(id2, 1, txO2)

		expectedUtxOs := make([]repository.UTxO, 0)
		expectedUtxOs = append(expectedUtxOs, uTxO1)
		expectedUtxOs = append(expectedUtxOs, uTxO2)

		// expectedUtxOs := []repository.UTxO{uTxO1, uTxO2}

		uTxOs, _, err := userWallet.FindUTxOs(100)
		if err != nil {
			t.Fatalf("expected to find UTxOs. err: %s", err)
		}

		if !reflect.DeepEqual(uTxOs, expectedUtxOs) {
			t.Fatalf("incorrect UtxOs\nGot:%+v\nWant:%+v", uTxOs, expectedUtxOs)
		}
	})

	test.Run("UTxOs cannot service the amount - insufficient funds", func(t *testing.T) {
		userWallet := wallet.NewWallet(*crypt)
		repository.ClearUTxOSet()

		id1 := []byte{1, 2, 3}
		id2 := []byte{4, 5, 6}

		txO1 := repository.TxO{
			Address: crypt.PublicKey,
			Amount:  70,
		}

		txO2 := repository.TxO{
			Address: crypt.PublicKey,
			Amount:  20,
		}

		repository.AddTxOToReceiver(id1, 1, txO1)
		repository.AddTxOToReceiver(id2, 1, txO2)

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
		userWallet := wallet.NewWallet(*crypt)

		id1 := []byte{1, 2, 3}
		uTxO1 := repository.UTxO{
			ID: repository.UTxOID{
				Address: crypt.PublicKey,
				TxID:    id1,
			},
			Amount: 70,
			Index:  1,
		}

		id2 := []byte{4, 5, 6}
		uTxO2 := repository.UTxO{
			ID: repository.UTxOID{
				Address: crypt.PublicKey,
				TxID:    id2,
			},
			Amount: 50,
			Index:  1,
		}

		uTxOs := make([]repository.UTxO, 0)
		uTxOs = append(uTxOs, uTxO1)
		uTxOs = append(uTxOs, uTxO2)

		receiverAddress := []byte{1, 2, 3}

		expectedTxOs := make([]repository.TxO, 0)
		expectedTxOs = append(expectedTxOs, repository.TxO{
			Address: receiverAddress,
			Amount:  100,
		})

		expectedTxOs = append(expectedTxOs, repository.TxO{
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
		userWallet := wallet.NewWallet(*crypt)

		id1 := []byte{1, 2, 3}
		uTxO1 := repository.UTxO{
			ID: repository.UTxOID{
				Address: crypt.PublicKey,
				TxID:    id1,
			},
			Amount: 20,
			Index:  1,
		}

		id2 := []byte{4, 5, 6}
		uTxO2 := repository.UTxO{
			ID: repository.UTxOID{
				Address: crypt.PublicKey,
				TxID:    id2,
			},
			Amount: 50,
			Index:  1,
		}

		uTxOs := make([]repository.UTxO, 0)
		uTxOs = append(uTxOs, uTxO1)
		uTxOs = append(uTxOs, uTxO2)

		receiverAddress := []byte{1, 2, 3}

		expectedTxOs := make([]repository.TxO, 0)
		expectedTxOs = append(expectedTxOs, repository.TxO{
			Address: receiverAddress,
			Amount:  100,
		})

		expectedTxOs = append(expectedTxOs, repository.TxO{
			Address: crypt.PublicKey,
			Amount:  20,
		})

		_, err := userWallet.GetTxOs(100, receiverAddress, uTxOs)
		if err == nil {
			t.Fatalf("expected error of insufficient funds")
		}

	})
}

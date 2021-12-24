package wallet_test

import (
	"blockchain/repository"
	"blockchain/wallet"
	"reflect"
	"sort"
	"testing"
)

func TestCreateCoinbaseTransaction(t *testing.T) {
	t.Run("validate successful coinbase transaction", func(t *testing.T) {
		crypt := wallet.NewCryptographic()
		crypt.GenerateKeyPair()

		coinbaseTx, now := wallet.CreateCoinbaseTransaction(*crypt, 0)

		txIns := make([]repository.TxIn, 0)
		txOuts := make([]repository.TxO, 0)

		txOut := repository.TxO{
			Value:        wallet.COINBASE_TRANSACTION_AMOUNT,
			ScriptPubKey: crypt.PublicKey,
		}
		txOuts = append(txOuts, txOut)

		expectedTx := repository.Transaction{
			ID:        []byte{},
			TxIns:     txIns,
			TxOuts:    txOuts,
			Timestamp: now,
		}

		txID := wallet.GenerateTransactionID(expectedTx)
		expectedTx.ID = txID

		if err := wallet.IsValidCoinbaseTransaction(coinbaseTx, []repository.Transaction{}); err != nil {
			t.Fatalf("coinbase Tx not valid %s", err.Error())
		}

		if !reflect.DeepEqual(expectedTx.Timestamp, coinbaseTx.Timestamp) {
			t.Fatalf("coinbase Tx timestamp not equal to expected Tx timestamp")
		}

		if !reflect.DeepEqual(expectedTx.TxOuts, coinbaseTx.TxOuts) {
			t.Fatalf("coinbase Tx TxOuts not equal to expected Tx TxOuts")
		}
	})
}

func TestCreateTransaction(t *testing.T) {
	t.Run("validate successful transaction - 2 inputs with change", func(t *testing.T) {
		senderCrypt := wallet.NewCryptographic()
		senderCrypt.GenerateKeyPair()

		receiverCrypt := wallet.NewCryptographic()
		receiverCrypt.GenerateKeyPair()

		amount := 8

		coinbaseTx, now := wallet.CreateCoinbaseTransaction(*senderCrypt, 0)
		repository.AddTxToUTxOSet(coinbaseTx)

		txIns := make([]repository.TxIn, 0)
		txOuts := make([]repository.TxO, 0)

		txIn := repository.TxIn{
			TxID:            coinbaseTx.ID,
			TxOIndex:        0,
			ScriptSignature: []byte{},
		}

		txIns = append(txIns, txIn)

		txOutReceiver := repository.TxO{
			ScriptPubKey: receiverCrypt.PublicKey,
			Value:        amount,
		}
		txOutSenderChange := repository.TxO{
			ScriptPubKey: senderCrypt.PublicKey,
			Value:        2,
		}
		txOuts = append(txOuts, txOutReceiver)
		txOuts = append(txOuts, txOutSenderChange)

		expectedSenderTx := repository.Transaction{
			ID:     []byte{},
			TxIns:  txIns,
			TxOuts: txOuts,
		}

		senderWallet := wallet.NewWallet(*senderCrypt)

		tx, now, _ := senderWallet.CreateTransaction(receiverCrypt.PublicKey, amount)

		expectedSenderTx.Timestamp = now

		expectedTxID := wallet.GenerateTransactionID(expectedSenderTx)

		expectedSenderTx.ID = expectedTxID

		txIns[0].ScriptSignature = senderCrypt.GenerateSignature(expectedTxID)

		if err := wallet.IsValidTransaction(*tx); err != nil {
			t.Fatalf("Test failed: %+v", err)
		}

		if !reflect.DeepEqual(expectedSenderTx.Timestamp, tx.Timestamp) {
			t.Fatalf("Tx timestamp not equal to expected Tx timestamp")
		}

		if !reflect.DeepEqual(len(expectedSenderTx.TxOuts), len(tx.TxOuts)) {
			t.Fatalf("Tx TxOuts not equal to expected Tx TxOuts")
		}

		if !reflect.DeepEqual(expectedSenderTx.TxIns[0].TxID, tx.TxIns[0].TxID) {
			t.Fatalf("Tx Ins UTxOID not equal to expected Tx Ins UTxOID")
		}

		if !reflect.DeepEqual(expectedSenderTx.TxIns[0].TxOIndex, tx.TxIns[0].TxOIndex) {
			t.Fatalf("Tx Ins UTxOIndex not equal to expected Tx Ins UTxOIndex")
		}
	})

	t.Run("invalidate bad tx - not enough money", func(t *testing.T) {
		senderCrypt := wallet.NewCryptographic()
		senderCrypt.GenerateKeyPair()

		receiverCrypt := wallet.NewCryptographic()
		receiverCrypt.GenerateKeyPair()

		amount := wallet.COINBASE_TRANSACTION_AMOUNT + 1

		coinbaseTx, _ := wallet.CreateCoinbaseTransaction(*senderCrypt, 0)
		repository.AddTxToUTxOSet(coinbaseTx)

		txIns := make([]repository.TxIn, 0)
		txOuts := make([]repository.TxO, 0)

		txIn := repository.TxIn{
			TxID:            coinbaseTx.ID,
			TxOIndex:        0,
			ScriptSignature: []byte{},
		}

		txIns = append(txIns, txIn)

		txOutReceiver := repository.TxO{
			ScriptPubKey: receiverCrypt.PublicKey,
			Value:        amount,
		}
		txOutSenderChange := repository.TxO{
			ScriptPubKey: senderCrypt.PublicKey,
			Value:        2,
		}
		txOuts = append(txOuts, txOutReceiver)
		txOuts = append(txOuts, txOutSenderChange)

		senderWallet := wallet.NewWallet(*senderCrypt)

		_, _, err := senderWallet.CreateTransaction(receiverCrypt.PublicKey, amount)
		if err == nil {
			t.Fatalf("Test failed: expected insufficient funds error")
		}
	})

	t.Run("invalidate bad tx - signature incorrect", func(t *testing.T) {
		senderCrypt := wallet.NewCryptographic()
		senderCrypt.GenerateKeyPair()

		receiverCrypt := wallet.NewCryptographic()
		receiverCrypt.GenerateKeyPair()

		amount := 8

		coinbaseTx, now := wallet.CreateCoinbaseTransaction(*senderCrypt, 0)
		repository.AddTxToUTxOSet(coinbaseTx)

		txIns := make([]repository.TxIn, 0)
		txOuts := make([]repository.TxO, 0)

		txIn := repository.TxIn{
			TxID:            coinbaseTx.ID,
			TxOIndex:        0,
			ScriptSignature: []byte{},
		}

		txIns = append(txIns, txIn)

		txOutReceiver := repository.TxO{
			ScriptPubKey: receiverCrypt.PublicKey,
			Value:        amount,
		}
		txOutSenderChange := repository.TxO{
			ScriptPubKey: senderCrypt.PublicKey,
			Value:        2,
		}
		txOuts = append(txOuts, txOutReceiver)
		txOuts = append(txOuts, txOutSenderChange)

		expectedSenderTx := repository.Transaction{
			ID:     []byte{},
			TxIns:  txIns,
			TxOuts: txOuts,
		}

		senderWallet := wallet.NewWallet(*senderCrypt)

		tx, now, _ := senderWallet.CreateTransaction(receiverCrypt.PublicKey, amount)

		expectedSenderTx.Timestamp = now

		expectedTxID := wallet.GenerateTransactionID(expectedSenderTx)

		expectedSenderTx.ID = expectedTxID

		tx.TxIns[0].ScriptSignature = []byte{1, 2, 3}

		if err := wallet.IsValidTransaction(*tx); err == nil ||
			(err != nil && err.Error() != "invalid txIn error in txIn number 0. error: Invalid transaction - signature verification failed: crypto/rsa: verification error") {
			t.Fatalf("Test failed: expected signature check to fail")
		}
	})
}

func TestFindUTxOs(test *testing.T) {
	test.Run("UTxOs can service the amount and fee", func(t *testing.T) {
		crypt := wallet.NewCryptographic()
		crypt.GenerateKeyPair()
		senderWallet := wallet.NewWallet(*crypt)

		crypt2 := wallet.NewCryptographic()
		crypt2.GenerateKeyPair()
		senderReceiverWallet := wallet.NewWallet(*crypt2)

		crypt3 := wallet.NewCryptographic()
		crypt3.GenerateKeyPair()

		coinbaseTx, _ := wallet.CreateCoinbaseTransaction(*crypt, 0)
		repository.AddTxToUTxOSet(coinbaseTx)

		tx, _, _ := senderWallet.CreateTransaction(crypt2.PublicKey, 5)
		tx2, _, _ := senderWallet.CreateTransaction(crypt2.PublicKey, 6)
		repository.AddTxToUTxOSet(*tx)
		repository.AddTxToUTxOSet(*tx2)

		uTxOs, _, err := senderReceiverWallet.FindUTxOs(10)
		if err != nil {
			t.Fatalf("expected to find UTxOs. err: %s", err)
		}

		expectedUtxOs := []wallet.TxIDIndexPair{
			{
				TxID:     tx2.ID,
				TxOIndex: 0,
			},
			{
				TxID:     tx.ID,
				TxOIndex: 0,
			},
		}

		sort.SliceStable(expectedUtxOs, func(i, j int) bool {
			return expectedUtxOs[i].TxID[0] < expectedUtxOs[j].TxID[0]
		})

		sort.SliceStable(uTxOs, func(i, j int) bool {
			return uTxOs[i].TxID[0] < uTxOs[j].TxID[0]
		})

		if !reflect.DeepEqual(uTxOs, expectedUtxOs) {
			t.Fatalf("incorrect UtxOs\nGot:%+v\nWant:%+v", uTxOs, expectedUtxOs)
		}
	})

	test.Run("UTxOs cannot service the amount - insufficient funds", func(t *testing.T) {
		crypt := wallet.NewCryptographic()
		crypt.GenerateKeyPair()
		senderWallet := wallet.NewWallet(*crypt)

		crypt2 := wallet.NewCryptographic()
		crypt2.GenerateKeyPair()
		senderReceiverWallet := wallet.NewWallet(*crypt2)

		coinbaseTx, _ := wallet.CreateCoinbaseTransaction(*crypt, 0)
		repository.AddTxToUTxOSet(coinbaseTx)

		tx, _, _ := senderWallet.CreateTransaction(crypt2.PublicKey, 5)
		tx2, _, _ := senderWallet.CreateTransaction(crypt2.PublicKey, 5)
		repository.AddTxToUTxOSet(*tx)
		repository.AddTxToUTxOSet(*tx2)

		_, _, err := senderReceiverWallet.FindUTxOs(11)
		if err == nil || (err != nil && err.Error() != "insufficient funds or no available uTxOs") {
			t.Fatalf("incorrect error type, expected insufficient funds, got: %+v", err)
		}
	})

	test.Run("UTxOs cannot service the amount - not enough for service fee", func(t *testing.T) {
		crypt := wallet.NewCryptographic()
		crypt.GenerateKeyPair()
		senderWallet := wallet.NewWallet(*crypt)

		crypt2 := wallet.NewCryptographic()
		crypt2.GenerateKeyPair()
		senderReceiverWallet := wallet.NewWallet(*crypt2)

		coinbaseTx, _ := wallet.CreateCoinbaseTransaction(*crypt, 0)
		repository.AddTxToUTxOSet(coinbaseTx)

		tx, _, _ := senderWallet.CreateTransaction(crypt2.PublicKey, 5)
		tx2, _, _ := senderWallet.CreateTransaction(crypt2.PublicKey, 5)
		repository.AddTxToUTxOSet(*tx)
		repository.AddTxToUTxOSet(*tx2)

		_, _, err := senderReceiverWallet.FindUTxOs(10)
		if err == nil || (err != nil && err.Error() != "insufficient funds to include the tx fee of 1 coin") {
			t.Fatalf("incorrect error type, expected insufficient funds, got: %+v", err)
		}
	})
}

func TestGetTxOs(test *testing.T) {
	crypt := wallet.NewCryptographic()
	crypt.GenerateKeyPair()

	test.Run("UTxOs can service the amount", func(t *testing.T) {
		amount := 6

		crypt := wallet.NewCryptographic()
		crypt.GenerateKeyPair()
		senderWallet := wallet.NewWallet(*crypt)

		crypt2 := wallet.NewCryptographic()
		crypt2.GenerateKeyPair()

		coinbaseTx, _ := wallet.CreateCoinbaseTransaction(*crypt, 0)
		repository.AddTxToUTxOSet(coinbaseTx)

		tx, _, _ := senderWallet.CreateTransaction(crypt2.PublicKey, 6)

		txOs := tx.TxOuts

		sort.SliceStable(txOs, func(i, j int) bool {
			return txOs[i].Value > txOs[j].Value
		})

		expectedTxO1 := repository.TxO{
			ScriptPubKey: crypt.PublicKey,
			Value:        wallet.COINBASE_TRANSACTION_AMOUNT - amount - wallet.TRANSACTION_FEE,
		}

		expectedTxO2 := repository.TxO{
			ScriptPubKey: crypt2.PublicKey,
			Value:        amount,
		}

		if !reflect.DeepEqual(txOs[0], expectedTxO1) {
			t.Fatalf("incorrect txO1 \nGot:%+v\nWant:%+v", txOs[0], expectedTxO1)
		}

		if !reflect.DeepEqual(txOs[1], expectedTxO2) {
			t.Fatalf("incorrect txO2 \nGot:%+v\nWant:%+v", txOs[1], expectedTxO2)
		}
	})
}

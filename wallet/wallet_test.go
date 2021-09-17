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
		txOuts := make([]wallet.TxOut, 0)
		txIn := wallet.TxIn{
			UTxOID: wallet.UTxOID{
				Address: []byte{},
				TxID:    []byte{},
			},
			UTxOIndex: 1,
			Signature: []byte{},
		}

		txIns = append(txIns, txIn)

		txOut := wallet.TxOut{
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

		expectedTxSigErr := crypt.VerifySignature(expectedTx.TxIns[0].Signature, crypt.PublicKey, wallet.GenerateTransactionID(expectedTx))
		coinbaseTxSigErr := crypt.VerifySignature(coinbaseTx.TxIns[0].Signature, crypt.PublicKey, wallet.GenerateTransactionID(coinbaseTx))

		if expectedTxSigErr != nil || coinbaseTxSigErr != nil {
			t.Fatalf("coinbase Tx Signatures invalid")
		}
	})
}

func TestCreateTransaction(t *testing.T) {
	t.Run("validate successful transaction - simple 1 input 1 output. Input perfectly adds up to output - no change", func(t *testing.T) {
		amount := 100

		// the output of a previous transaction that was sent to the sender of this transaction.
		previousTxID := []byte{1, 2, 3}

		senderCrypt := wallet.NewCryptographic()
		senderCrypt.GenerateKeyPair()

		receiverCrypt := wallet.NewCryptographic()
		receiverCrypt.GenerateKeyPair()

		uTxOSet := make(map[wallet.PublicKeyAddressType]map[wallet.TxIDType]wallet.UTxO, 0)

		txIns := make([]wallet.TxIn, 0)
		txOuts := make([]wallet.TxOut, 0)

		txIn := wallet.TxIn{
			UTxOID: wallet.UTxOID{
				Address: senderCrypt.PublicKey,
				TxID:    previousTxID,
			},
			UTxOIndex: 1,
			Signature: []byte{},
		}
		txIns = append(txIns, txIn)

		txOutReceiver := wallet.TxOut{
			Address: receiverCrypt.PublicKey,
			Amount:  amount,
		}
		txOuts = append(txOuts, txOutReceiver)

		expectedSenderTx := wallet.Transaction{
			ID:     []byte{},
			TxIns:  txIns,
			TxOuts: txOuts,
		}

		uTxO := wallet.UTxO{
			ID: wallet.UTxOID{
				TxID:    previousTxID,
				Address: senderCrypt.PublicKey,
			},
			Index:  1,
			Amount: 200,
		}

		tx, now := wallet.CreateTransaction(receiverCrypt.PublicKey, amount, &uTxO, senderCrypt)
		expectedSenderTx.Timestamp = now

		expectedTxID := wallet.GenerateTransactionID(expectedSenderTx)
		// uTxO.ID = expectedTxID

		uTxOTxIDMap := make(map[wallet.TxIDType]wallet.UTxO)
		uTxOTxIDMap[wallet.TxIDType(previousTxID)] = uTxO
		uTxOSet[wallet.PublicKeyAddressType(senderCrypt.PublicKey)] = uTxOTxIDMap

		senderWallet := wallet.NewWallet(uTxOSet, *senderCrypt)

		// this should actually be the txID of a different, older, tx.
		expectedSenderTx.ID = expectedTxID

		txIns[0].UTxOID.TxID = previousTxID
		txIns[0].Signature = senderCrypt.GenerateSignature(expectedTxID)

		if err := senderWallet.IsValidTransaction(expectedSenderTx); err != nil {
			t.Fatalf("Test failed: %+v", err)
		}

		if !reflect.DeepEqual(expectedSenderTx.Timestamp, tx.Timestamp) {
			t.Fatalf("Tx timestamp not equal to expected Tx timestamp")
		}

		if !reflect.DeepEqual(expectedSenderTx.TxOuts, tx.TxOuts) {
			t.Fatalf("Tx TxOuts not equal to expected Tx TxOuts")
		}

		if !reflect.DeepEqual(expectedSenderTx.TxIns[0].UTxOID, tx.TxIns[0].UTxOID) {
			t.Fatalf("Tx Ins UTxOID not equal to expected Tx Ins UTxOID")
		}

		if !reflect.DeepEqual(expectedSenderTx.TxIns[0].UTxOIndex, tx.TxIns[0].UTxOIndex) {
			t.Fatalf("Tx Ins UTxOIndex not equal to expected Tx Ins UTxOIndex")
		}
	})
}

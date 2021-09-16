package wallet_test

import (
	"blockchain/wallet"
	"reflect"
	"testing"
	"time"
)

func TestCreateCoinbaseTransaction(t *testing.T) {
	t.Run("validate successful coinbase transaction", func(t *testing.T) {
		account := wallet.NewAccount()
		account.GenerateKeyPair()

		coinbaseTx, now := wallet.CreateCoinbaseTransaction(*account, 1)

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
			Address: account.PublicKey,
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

		txInSignature := account.GenerateSignature(txID)
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

		isValidExpectedTxSig, _ := account.VerifySignature(expectedTx.TxIns[0].Signature, account.PublicKey, wallet.GenerateTransactionID(expectedTx))
		isValidCoinbaseTxSig, _ := account.VerifySignature(coinbaseTx.TxIns[0].Signature, account.PublicKey, wallet.GenerateTransactionID(coinbaseTx))

		if !isValidExpectedTxSig || !isValidCoinbaseTxSig {
			t.Fatalf("coinbase Tx Signatures invalid")
		}
	})
}

func TestCreateTransaction(t *testing.T) {
	t.Run("validate successful transaction - simple 1 input 1 output. Input perfectly adds up to output - no change", func(t *testing.T) {

		senderAccount := wallet.NewAccount()
		senderAccount.GenerateKeyPair()

		receiverAccount := wallet.NewAccount()
		receiverAccount.GenerateKeyPair()

		uTxOSet := make(map[wallet.PublicKeyAddressType]map[wallet.TxIDType]wallet.UTxO, 0)

		txIns := make([]wallet.TxIn, 0)
		txOuts := make([]wallet.TxOut, 0)

		txIn := wallet.TxIn{
			UTxOID: wallet.UTxOID{
				Address: senderAccount.PublicKey,
				TxID:    []byte{},
			},
			UTxOIndex: 1,
			Signature: []byte{},
		}
		txIns = append(txIns, txIn)

		txOutReceiver := wallet.TxOut{
			Address: receiverAccount.PublicKey,
			Amount:  100,
		}
		txOuts = append(txOuts, txOutReceiver)

		expectedSenderTransaction := wallet.Transaction{
			ID:        []byte{},
			TxIns:     txIns,
			TxOuts:    txOuts,
			Timestamp: int(time.Now().UnixNano()),
		}

		txID := wallet.GenerateTransactionID(expectedSenderTransaction)

		uTxO := wallet.UTxO{
			ID:      txID,
			Index:   1,
			Address: senderAccount.PublicKey,
			Amount:  200,
		}

		uTxOTxIDMap := make(map[wallet.TxIDType]wallet.UTxO)
		uTxOTxIDMap[wallet.TxIDType(txID)] = uTxO
		uTxOSet[wallet.PublicKeyAddressType(senderAccount.PublicKey)] = uTxOTxIDMap
		senderWallet := wallet.NewWallet(uTxOSet)

		expectedSenderTransaction.ID = txID

		txIns[0].UTxOID.TxID = txID
		txIns[0].Signature = senderAccount.GenerateSignature(txID)

		if !senderWallet.IsValidTransaction(expectedSenderTransaction) {
			t.Fatalf("expected transaction incorrectly constructed")
		}
	})
}

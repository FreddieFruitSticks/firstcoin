package wallet_test

import (
	"blockchain/wallet"
	"reflect"
	"testing"
)

func TestCreateCoinbaseTransaction(t *testing.T) {
	t.Run("validate successful coinbase transaction", func(t *testing.T) {
		account := wallet.NewAccount()
		account.GenerateKeyPair()

		coinbaseTx, now := wallet.CreateCoinbaseTransaction(*account, 1)

		txIns := make([]wallet.TxIn, 0)
		txOuts := make([]wallet.TxOut, 0)
		txIn := wallet.TxIn{
			UTxOID:    []byte{},
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
	t.Run("validate successful transaction", func(t *testing.T) {
		account := wallet.NewAccount()
		account.GenerateKeyPair()

		// expectedTransaction := wallet.Transaction{

		// }
	})
}

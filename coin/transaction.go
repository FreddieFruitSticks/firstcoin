package coin

import "fmt"

const COINBASE_TRANSACTION = 10

type Transaction struct {
	ID    []byte `json:"id"`
	TxIn  TxIn   `json:"transactionInput"`
	TxOut TxOut  `json:"transactionOutput"`
}

// transcation input refers to the giver of coins. Signature is signed with giver's private key
type TxIn struct {
	ReferenceUnspentTxOutID    string
	ReferenceUnspentTxOutIndex int
	Signature                  []byte
}

// transcation ouput refers to the receiver of coins. Address is receiver's public key
type TxOut struct {
	Address []byte
	Amount  int
}

type UnspentTxOut struct {
	ID      string
	Index   int
	Address []byte
	Amount  int
}

func CreateNewTransactionPool(address []byte, account Account) []Transaction {
	transactionPool := make([]Transaction, 0)
	coinbaseTransaction := CreateCoinbaseTransaction(address, account)

	transactionPool = append(transactionPool, coinbaseTransaction)

	return transactionPool
}

func CreateTransaction(address []byte, amount int, refUnspentTxOut *UnspentTxOut, a *Account) Transaction {
	txIn := TxIn{
		ReferenceUnspentTxOutID:    refUnspentTxOut.ID,
		ReferenceUnspentTxOutIndex: refUnspentTxOut.Index,
	}

	txOut := TxOut{
		Amount:  COINBASE_TRANSACTION,
		Address: address,
	}

	txSignature := generateTransactionID(address, COINBASE_TRANSACTION, txIn.ReferenceUnspentTxOutID, txIn.ReferenceUnspentTxOutIndex)

	signature := a.GenerateSignature(txSignature)
	txIn.Signature = signature

	return Transaction{
		ID:    txSignature,
		TxIn:  txIn,
		TxOut: txOut,
	}
}

func CreateCoinbaseTransaction(address []byte, a Account) Transaction {
	txIn := TxIn{
		ReferenceUnspentTxOutID:    "",
		ReferenceUnspentTxOutIndex: 0,
	}

	txOut := TxOut{
		Amount:  COINBASE_TRANSACTION,
		Address: address,
	}

	txSignature := generateTransactionID(address, COINBASE_TRANSACTION, txIn.ReferenceUnspentTxOutID, txIn.ReferenceUnspentTxOutIndex)

	signature := a.GenerateSignature(txSignature)
	txIn.Signature = signature

	return Transaction{
		ID:    txSignature,
		TxIn:  txIn,
		TxOut: txOut,
	}
}

func (t Transaction) String() string {
	return fmt.Sprintf("uTxOID: %s,\nuTxOIndex: %d,\ntxOAddress: %s,\ntxOAmount: %d", t.TxIn.ReferenceUnspentTxOutID, t.TxIn.ReferenceUnspentTxOutIndex, t.TxOut.Address, t.TxOut.Amount)
}

package wallet

import (
	"blockchain/utils"
	"crypto/sha256"
	"fmt"
	"strconv"
)

const COINBASE_TRANSACTION = 10

type Transaction struct {
	ID     []byte  `json:"id"`
	TxIns  []TxIn  `json:"transactionInputs"`
	TxOuts []TxOut `json:"transactionOutputs"`
}

// transcation input refers to the giver of coins. Signature is signed with giver's private key
type TxIn struct {
	UTxOID    []byte
	UTxOIndex int // index is the block number or block height - this is to prevent duplicate signatures for exact same txs
	Signature []byte
}

// transcation ouput refers to the receiver of coins. Address is receiver's public key
type TxOut struct {
	Address []byte
	Amount  int
}

// unspend transaction outputs are the final transaction outs allocated to each key - it's ID is the ID of the transaction that created it
type UTxOut struct {
	ID      []byte
	Index   int
	Address []byte
	Amount  int
}

func CreateNewTransactionPool(address []byte, account Account) []Transaction {
	transactionPool := make([]Transaction, 0)

	// coinbase transaction is the first transaction included by the miner
	coinbaseTransaction := CreateCoinbaseTransaction(address, account)

	transactionPool = append(transactionPool, coinbaseTransaction)

	return transactionPool
}

func CreateTransaction(address []byte, amount int, uTxO *UTxOut, a *Account) Transaction {
	txIns := make([]TxIn, 0)
	txOuts := make([]TxOut, 0)

	txIn := TxIn{
		UTxOID:    uTxO.ID,
		UTxOIndex: uTxO.Index,
	}
	txIns = append(txIns, txIn)

	txOut := TxOut{
		Amount:  COINBASE_TRANSACTION,
		Address: address,
	}
	txOuts = append(txOuts, txOut)

	transaction := Transaction{
		TxIns:  txIns,
		TxOuts: txOuts,
	}

	// Generate the transaction id from the tx inputs (without signature) and tx outputs
	txSignature := generateTransactionID(transaction)
	transaction.ID = txSignature

	// tx input signature is the tx id signed by the spender of coins
	signature := a.GenerateSignature(txSignature)
	txIn.Signature = signature

	return Transaction{
		ID:     txSignature,
		TxIns:  txIns,
		TxOuts: txOuts,
	}
}

func CreateCoinbaseTransaction(address []byte, a Account) Transaction {
	// First create the transaction with TxIns and TxOuts - tx id and txIn signature are not included yet
	txIns := make([]TxIn, 0)
	txOuts := make([]TxOut, 0)
	txIn := TxIn{
		UTxOID:    []byte{},
		UTxOIndex: 0,
	}

	txIns = append(txIns, txIn)

	txOut := TxOut{
		Amount:  COINBASE_TRANSACTION,
		Address: address,
	}
	txOuts = append(txOuts, txOut)

	transaction := Transaction{
		TxIns:  txIns,
		TxOuts: txOuts,
	}

	// Generate the transaction id from the tx inputs (without signature) and tx outputs
	txSignature := generateTransactionID(transaction)
	transaction.ID = txSignature

	// tx input signature is the tx id signed by the spender of coins - this will be used later to verify that the uTxO public key
	// is the true owner and authoriser of this signed transaction
	txInSignature := a.GenerateSignature(txSignature)
	txIns[0].Signature = txInSignature

	return transaction
}

// This is a SHA of all txIns (excluding signature - that gets added later) and txOuts
func generateTransactionID(transaction Transaction) []byte {
	msgHash := sha256.New()
	concatTxIn := ""
	concatTxOut := ""

	for _, txIn := range transaction.TxIns {
		concatTxIn += string(txIn.UTxOID) + strconv.Itoa(txIn.UTxOIndex)
	}

	for _, txOut := range transaction.TxOuts {
		concatTxOut += string(txOut.Address) + strconv.Itoa(txOut.Amount)
	}

	_, err := msgHash.Write([]byte(fmt.Sprintf("%s%s", concatTxIn, concatTxOut)))
	utils.CheckError(err)

	return msgHash.Sum(nil)
}

func (t Transaction) String() string {
	return fmt.Sprintf("txId: %s\ntxIns: %+v\ntxOuts: %+v", t.ID, t.TxIns, t.TxOuts)
}

type prettyTxO struct {
	Address string
	Amount  int
}

// func (t *TxOut) MarshalJSON() ([]byte, error) {
// 	prettyTxO := prettyTxO{
// 		Address: string(t.Address),
// 		Amount:  t.Amount,
// 	}
// 	marshal, err := json.Marshal(prettyTxO)

// 	return marshal, err
// }

package wallet

import (
	"blockchain/utils"
	"crypto/sha256"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

const COINBASE_TRANSACTION_AMOUNT = 10

type Transaction struct {
	ID        []byte  `json:"id"`
	TxIns     []TxIn  `json:"transactionInputs"`
	TxOuts    []TxOut `json:"transactionOutputs"`
	Timestamp int     `json:"timestamp"`
}

// transcation input refers to the giver of coins. Signature is signed with giver's private key
type TxIn struct {
	UTxOID    []byte
	UTxOIndex int // index is the block number or block height - this is to prevent duplicate signatures for exact same txs
	Signature []byte
}

// transaction ouput refers to the receiver of coins. Address is receiver's public key
type TxOut struct {
	Address []byte
	Amount  int
}

// unspend transaction outputs are the final transaction outs allocated to each key - it's ID is the ID of the transaction that created it
type UTxOut struct {
	ID      []byte //ID is unique because of timestamp
	Index   int
	Address []byte
	Amount  int
}

func CreateNewTransactionPool(account Account) []Transaction {
	transactionPool := make([]Transaction, 0)

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
		Amount:  COINBASE_TRANSACTION_AMOUNT,
		Address: address,
	}
	txOuts = append(txOuts, txOut)

	transaction := Transaction{
		TxIns:  txIns,
		TxOuts: txOuts,
	}

	now := int(time.Now().UnixNano())

	// Generate the transaction id from the tx inputs (without signature) and tx outputs
	txID := GenerateTransactionID(transaction)
	transaction.ID = txID

	// tx input signature is the tx id signed by the spender of coins
	signature := a.GenerateSignature(txID)
	txIn.Signature = signature

	return Transaction{
		ID:        txID,
		TxIns:     txIns,
		TxOuts:    txOuts,
		Timestamp: now,
	}
}

func CreateCoinbaseTransaction(a Account, blockIndex int) (Transaction, int) {
	// First create the transaction with TxIns and TxOuts - tx id and txIn signature are not included yet
	txIns := make([]TxIn, 0)
	txOuts := make([]TxOut, 0)
	txIn := TxIn{
		UTxOID:    []byte{},
		UTxOIndex: blockIndex,
	}

	txIns = append(txIns, txIn)

	txOut := TxOut{
		Amount:  COINBASE_TRANSACTION_AMOUNT,
		Address: a.PublicKey,
	}
	txOuts = append(txOuts, txOut)

	now := int(time.Now().UnixNano())

	transaction := Transaction{
		TxIns:     txIns,
		TxOuts:    txOuts,
		Timestamp: now,
	}

	// Generate the transaction id from the tx inputs (without signature) and tx outputs
	txID := GenerateTransactionID(transaction)
	transaction.ID = txID

	// tx input signature is the tx id signed by the spender of coins - this will be used later to verify that the uTxO public key
	// is the true owner and authoriser of this signed transaction
	txInSignature := a.GenerateSignature(txID)
	txIns[0].Signature = txInSignature

	return transaction, now
}

// This is a SHA of all txIns (excluding signature - that gets added later) and txOuts
func GenerateTransactionID(transaction Transaction) []byte {
	msgHash := sha256.New()
	concatTxIn := ""
	concatTxOut := ""

	for _, txIn := range transaction.TxIns {
		concatTxIn += string(txIn.UTxOID) + strconv.Itoa(txIn.UTxOIndex)
	}

	for _, txOut := range transaction.TxOuts {
		concatTxOut += string(txOut.Address) + strconv.Itoa(txOut.Amount)
	}

	_, err := msgHash.Write([]byte(fmt.Sprintf("%s%s%d", concatTxIn, concatTxOut, transaction.Timestamp)))
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

func AreValidTransactions(transactions []Transaction, blockIndex int) bool {
	if len(transactions) == 0 {
		return false
	}

	// first transaction in the list is always the coinbase transaction
	coinbaseTransaction := transactions[0]

	if !IsValidCoinbaseTransaction(coinbaseTransaction, blockIndex) {
		return false
	}

	// for _, transaction := range block.Transactions[1:] {

	// }

	return true
}

func IsValidTransaction(transaction Transaction, blockIndex int) bool {

	tID := GenerateTransactionID(transaction)
	if !reflect.DeepEqual(tID, transaction.ID) {
		return false
	}

	return true
}

func IsValidCoinbaseTransaction(transaction Transaction, blockIndex int) bool {
	if len(transaction.TxIns) != 1 {
		fmt.Println("Invalid coinbase transaction txIns length > 0")
		return false
	}

	if len(transaction.TxOuts) != 1 {
		fmt.Println("Invalid coinbase transaction txOuts length > 0")
		return false
	}

	if transaction.TxOuts[0].Amount != COINBASE_TRANSACTION_AMOUNT {
		fmt.Println("Invalid coinbase transaction amount != COINBASE_TRANSACTION_AMOUNT")
		return false
	}

	if transaction.TxIns[0].UTxOIndex != blockIndex {
		fmt.Println("Invalid coinbase transaction amount != COINBASE_TRANSACTION_AMOUNT")
		return false
	}

	tID := GenerateTransactionID(transaction)
	if !reflect.DeepEqual(tID, transaction.ID) {
		return false
	}

	return true
}

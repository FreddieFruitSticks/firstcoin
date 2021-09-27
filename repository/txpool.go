package repository

import "fmt"

type Transaction struct {
	ID        []byte `json:"id"`
	TxIns     []TxIn `json:"transactionInputs"`
	TxOuts    []TxO  `json:"transactionOutputs"`
	Timestamp int    `json:"timestamp"`
}

var unconfirmedTransactionPool = make([]Transaction, 0)

func (t Transaction) String() string {
	return fmt.Sprintf("txId: %s\ntxIns: %+v\ntxOuts: %+v", t.ID, t.TxIns, t.TxOuts)
}

func AddTxToTxPool(txs ...Transaction) {
	unconfirmedTransactionPool = append(unconfirmedTransactionPool, txs...)
}

func GetTxPool() []Transaction {
	return unconfirmedTransactionPool
}

func EmptyTxPool() {
	unconfirmedTransactionPool = make([]Transaction, 0)
}

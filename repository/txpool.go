package repository

import "fmt"

type Transaction struct {
	ID        []byte `json:"id"`
	TxIns     []TxIn `json:"transactionInputs"`
	TxOuts    []TxO  `json:"transactionOutputs"`
	Timestamp int    `json:"timestamp"`
}

var unconfirmedTransactionPool = make(map[TxIDType]Transaction, 0)

func (t Transaction) String() string {
	return fmt.Sprintf("txId: %s\ntxIns: %+v\ntxOuts: %+v\n", t.ID, t.TxIns, t.TxOuts)
}

func AddTxToTxPool(txs ...Transaction) {
	for _, tx := range txs {
		unconfirmedTransactionPool[TxIDType(tx.ID)] = tx
	}
}

func GetTxPool() map[TxIDType]Transaction {
	return unconfirmedTransactionPool
}

func GetTxFromTxPool(txId []byte) (Transaction, bool) {
	tx, ok := unconfirmedTransactionPool[TxIDType(txId)]
	return tx, ok
}

func GetTxPoolArray() []Transaction {
	txPool := make([]Transaction, 0)

	for _, tx := range unconfirmedTransactionPool {
		txPool = append(txPool, tx)
	}

	return txPool
}

func SetTxPool(txPool map[TxIDType]Transaction) {
	unconfirmedTransactionPool = txPool
}

func RemoveTxFromTxPool(txId []byte) {
	delete(unconfirmedTransactionPool, TxIDType(txId))
}

func EmptyTxPool() {
	unconfirmedTransactionPool = make(map[TxIDType]Transaction, 0)
}

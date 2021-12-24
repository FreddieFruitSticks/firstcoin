package wallet

import (
	"blockchain/repository"
	"blockchain/utils"
	"crypto/sha256"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

const COINBASE_TRANSACTION_AMOUNT = 10

type Wallet struct {
	Crypt Cryptographic
}

func NewWallet(c Cryptographic) *Wallet {
	return &Wallet{
		Crypt: c,
	}
}

func (w *Wallet) CreateTransaction(receiverAddress []byte, amount int) (*repository.Transaction, int, error) {
	txIns := make([]repository.TxIn, 0)
	txOuts := make([]repository.TxO, 0)

	txIDIndexPairs, _, err := w.FindUTxOs(amount)
	if err != nil {
		return nil, 0, err
	}

	for _, txIDIndexPair := range txIDIndexPairs {
		txIn := repository.TxIn{
			TxOIndex: txIDIndexPair.TxOIndex,
			TxID:     txIDIndexPair.TxID,
		}
		txIns = append(txIns, txIn)
	}

	txOs, _ := w.GetTxOs(amount, receiverAddress, txIns)
	for _, txO := range txOs {
		txOuts = append(txOuts, txO)
	}

	transaction := repository.Transaction{
		TxIns:  txIns,
		TxOuts: txOuts,
	}

	now := int(time.Now().UnixNano())
	transaction.Timestamp = now

	// Generate the transaction id from the tx inputs (without signature) and tx outputs
	txID := GenerateTransactionID(transaction)
	transaction.ID = txID

	// tx input signature is the tx id signed by the spender of coins
	signature := w.Crypt.GenerateSignature(txID)

	// sign all txIns from same sender.
	for i := 0; i < len(transaction.TxIns); i++ {
		transaction.TxIns[i].ScriptSignature = signature
	}

	return &transaction, now, nil
}

func CreateCoinbaseTransaction(crypt Cryptographic) (repository.Transaction, int) {
	// First create the transaction with TxIns and TxOuts - tx id and txIn signature are not included yet
	txIns := make([]repository.TxIn, 0)
	txOuts := make([]repository.TxO, 0)

	txOut := repository.TxO{
		Value:        COINBASE_TRANSACTION_AMOUNT,
		ScriptPubKey: crypt.PublicKey,
	}
	txOuts = append(txOuts, txOut)

	now := int(time.Now().UnixNano())

	transaction := repository.Transaction{
		TxIns:     txIns,
		TxOuts:    txOuts,
		Timestamp: now,
	}

	// Generate the transaction id from the tx inputs (without signature) and tx outputs
	txID := GenerateTransactionID(transaction)
	transaction.ID = txID

	return transaction, now
}

// This is a SHA of all txIns (excluding signature - that gets added later) and txOuts
func GenerateTransactionID(transaction repository.Transaction) []byte {
	msgHash := sha256.New()
	concatTxIn := ""
	concatTxOut := ""

	for _, txIn := range transaction.TxIns {
		concatTxIn += string(txIn.TxID) + strconv.Itoa(txIn.TxOIndex)
	}

	for _, txOut := range transaction.TxOuts {
		concatTxOut += string(txOut.ScriptPubKey) + strconv.Itoa(txOut.Value)
	}

	_, err := msgHash.Write([]byte(fmt.Sprintf("%s%s%d", concatTxIn, concatTxOut, transaction.Timestamp)))
	utils.CheckError(err)

	return msgHash.Sum(nil)
}

type prettyTxO struct {
	Address string
	Amount  int
}

func AreValidTransactions(transactions []repository.Transaction) error {
	if len(transactions) == 0 {
		return fmt.Errorf("Invalid transactions. Cant have empty transactions")
	}

	// first transaction in the list is always the coinbase transaction
	coinbaseTransaction := transactions[0]

	if err := IsValidCoinbaseTransaction(coinbaseTransaction); err != nil {
		return err
	}

	for _, transaction := range transactions[1:] {
		if err := IsValidTransaction(transaction); err != nil {
			return err
		}
	}

	return nil
}

func IsValidTransaction(tx repository.Transaction) error {
	return IsValidTransactionCopy(tx, repository.GetEntireUTxOSet())
}

func IsValidTransactionCopy(tx repository.Transaction, uTxOSet repository.UTxOSetType) error {
	if len(tx.TxIns) < 1 {
		return fmt.Errorf("Invalid transaction: txIns length must be greater than 0")
	}

	if err := AreValidTxIns(tx, uTxOSet); err != nil {
		return fmt.Errorf("invalid txIn %+v", err)
	}

	if len(tx.TxOuts) < 1 {
		return fmt.Errorf("Invalid transaction: txOuts length must be greater than 0")
	}

	if err := AreValidTxOuts(tx.TxOuts); err != nil {
		return fmt.Errorf("Invalid transaction, invalid txOut %+v", err)
	}

	tID := GenerateTransactionID(tx)
	if !reflect.DeepEqual(tID, tx.ID) {
		return fmt.Errorf("Invalid transaction id: %s. generated: %s", repository.Base64Encode(tx.ID), repository.Base64Encode(tID))
	}

	if err := VerifyTransactionAmountCopy(tx, uTxOSet); err != nil {
		return fmt.Errorf("Invalid transaction - amount verification failed: %+v", err.Error())
	}

	return nil
}

func IsValidCoinbaseTransaction(transaction repository.Transaction) error {
	if len(transaction.TxOuts) != 1 {
		return fmt.Errorf("Invalid coinbase transaction txOuts length > 0")
	}

	if transaction.TxOuts[0].Value != COINBASE_TRANSACTION_AMOUNT {
		return fmt.Errorf("Invalid coinbase transaction amount != COINBASE_TRANSACTION_AMOUNT")
	}

	tID := GenerateTransactionID(transaction)
	if !reflect.DeepEqual(tID, transaction.ID) {
		return fmt.Errorf("Invalid coinbase transaction. TransactionId is invalid")
	}

	return nil
}

func VerifyTransactionAmount(tx repository.Transaction) error {
	return VerifyTransactionAmountCopy(tx, repository.GetEntireUTxOSet())
}

func VerifyTransactionAmountCopy(tx repository.Transaction, uTxOSet repository.UTxOSetType) error {
	totalAmountFromUTxOs := 0

	for _, txIn := range tx.TxIns {
		spenderUTxO, err := getUTxOFromTxIn(txIn, uTxOSet)
		if err != nil {
			return err
		}

		if err := IsValidTxOutStructure(*spenderUTxO); err != nil {
			return err
		}
		totalAmountFromUTxOs += spenderUTxO.Value
	}

	if tx.TxOuts[0].Value > totalAmountFromUTxOs {
		return fmt.Errorf("unspent transaction output does not have enough coin")
	}

	return nil
}

func IsValidTxIn(txIn repository.TxIn, uTxOSet repository.UTxOSetType, txID []byte) error {
	spenderTx, ok := uTxOSet[repository.TxIDType(txIn.TxID)]
	if !ok {
		return fmt.Errorf("Invalid txIn - no tx for in set for txID in txIn")
	}

	if txIn.TxOIndex >= len(spenderTx.TxOuts) {
		return fmt.Errorf("Invalid txIn - referenced txO index does not exist")
	}

	uTxO := spenderTx.TxOuts[txIn.TxOIndex]
	if err := VerifySignature(txIn.ScriptSignature, uTxO.ScriptPubKey, txID); err != nil {
		return fmt.Errorf("Invalid transaction - signature verification failed: %+v", err.Error())

	}

	if len(txIn.TxID) == 0 {
		return fmt.Errorf("txIn UTxO TxID cannot be empty")
	}

	if len(txIn.ScriptSignature) == 0 {
		return fmt.Errorf("txIn Signature cannot be empty")
	}

	return nil
}

func getUTxOFromTxIn(txIn repository.TxIn, uTxOSet repository.UTxOSetType) (*repository.TxO, error) {
	spenderTx, ok := uTxOSet[repository.TxIDType(txIn.TxID)]
	if !ok {
		return nil, fmt.Errorf("Invalid txIn - no tx for in set for txID in txIn")
	}

	if txIn.TxOIndex >= len(spenderTx.TxOuts) {
		return nil, fmt.Errorf("Invalid txIn - referenced txO index does not exist")
	}

	return &spenderTx.TxOuts[txIn.TxOIndex], nil
}

func AreValidTxIns(tx repository.Transaction, uTxOSet repository.UTxOSetType) error {
	for index, txIn := range tx.TxIns {
		if err := IsValidTxIn(txIn, uTxOSet, tx.ID); err != nil {
			return fmt.Errorf("error in txIn number %d. error: %+v", index, err)
		}
	}

	return nil
}

func IsValidTxOutStructure(txOut repository.TxO) error {
	if len(txOut.ScriptPubKey) == 0 {
		return fmt.Errorf("invalid txOut: address must be valid public key")
	}

	if txOut.Value <= 0 {
		return fmt.Errorf("invalid txOut: amount must be > 0")
	}

	return nil
}

func AreValidTxOuts(txOuts []repository.TxO) error {
	for index, txOut := range txOuts {
		if err := IsValidTxOutStructure(txOut); err != nil {
			return fmt.Errorf("invalid txOut number %d. error: %+v", index, err)
		}
	}

	return nil
}

type TxIDIndexPair struct {
	TxID     []byte
	TxOIndex int
}

// finding the senders UTxOs that can service the Tx amount - currently the strategy is simply to take the first set of uTxOs
// that is not already included in the txPool. If the tx is in the txPool, we will alter the txOs and the index will not be correct for
// the next txIn - TODO: think of something smarter
func (w *Wallet) FindUTxOs(amount int) ([]TxIDIndexPair, int, error) {
	spenderLedger := repository.GetUserLedger(w.Crypt.PublicKey)
	uTxOs := make([]TxIDIndexPair, 0)

	totalAmount := 0

	for _, tx := range spenderLedger {
		for index, uTxO := range tx.TxOuts {
			if totalAmount < amount && !isUTxOInTxPool(tx.ID) && uTxOBelongsToSpender(uTxO, w.Crypt.PublicKey) {
				uTxOs = append(uTxOs, TxIDIndexPair{
					TxID:     tx.ID,
					TxOIndex: index,
				})
				totalAmount += uTxO.Value
				break

			}

			if totalAmount >= amount {
				return uTxOs, totalAmount, nil
			}
		}
	}

	if totalAmount < amount {
		return nil, totalAmount, fmt.Errorf("insufficient funds or no available uTxOs")
	}

	return uTxOs, totalAmount, nil
}

func uTxOBelongsToSpender(uTxO repository.TxO, spenderPublicKey []byte) bool {
	return reflect.DeepEqual(uTxO.ScriptPubKey, spenderPublicKey)
}

// can only send to one receiver, and can get change
func (w *Wallet) GetTxOs(amount int, receiverAddress []byte, txIns []repository.TxIn) ([]repository.TxO, error) {
	txOs := make([]repository.TxO, 0)

	valid, totalAmount := w.validateTxInsCanServiceAmount(txIns, amount)
	if !valid {
		return nil, fmt.Errorf("uTxOs cannot service amount. uTxO total: %d. amount: %d", totalAmount, amount)
	}

	txO := repository.TxO{
		ScriptPubKey: receiverAddress,
		Value:        amount,
	}

	txOs = append(txOs, txO)

	change := 0
	// deduct the difference between the total amount and the amount required, and add that as a repository.TxO to go back to the spender (as change)
	if totalAmount > amount {
		change = totalAmount - amount
		changeTxO := repository.TxO{
			ScriptPubKey: w.Crypt.PublicKey,
			Value:        change,
		}

		txOs = append(txOs, changeTxO)
	}

	return txOs, nil
}

func (w *Wallet) validateTxInsCanServiceAmount(txIns []repository.TxIn, amount int) (bool, int) {
	totalAmount := 0

	for _, txIn := range txIns {
		spenderTxs := repository.GetUserLedger(w.Crypt.PublicKey)
		tx := spenderTxs[repository.TxIDType(txIn.TxID)]
		totalAmount += tx.TxOuts[txIn.TxOIndex].Value
	}

	return totalAmount >= amount, totalAmount
}

func isUTxOInTxPool(txID []byte) bool {
	txPool := repository.GetTxPool()

	for _, tx := range txPool {
		for _, txIn := range tx.TxIns {
			if reflect.DeepEqual(txIn.TxID, txID) {
				return true
			}
		}
	}

	return false
}

func GetTotalAmount(publicKey []byte) int {
	userLedger := repository.GetUserLedger(publicKey)

	totalAmount := 0

	for _, tx := range userLedger {
		for _, uTxO := range tx.TxOuts {
			if reflect.DeepEqual(uTxO.ScriptPubKey, publicKey) {
				totalAmount += uTxO.Value
			}
		}
	}

	return totalAmount
}

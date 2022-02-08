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

const COINBASE_TRANSACTION_AMOUNT = 100
const TRANSACTION_FEE = 1

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
	sigScript := w.GenerateTxSigScript(txID)

	// sign all txIns from same sender.
	for i := 0; i < len(transaction.TxIns); i++ {
		transaction.TxIns[i].ScriptSignature = sigScript
	}

	return &transaction, now, nil
}

func (w *Wallet) GenerateTxSigScript(txID []byte) []byte {
	signature := w.Crypt.GenerateSignature(txID)
	publicKey := w.Crypt.PublicKey

	sigScript := append(signature, []byte(fmt.Sprintf("[%s]", sigHashAll))...)
	sigScript = append(sigScript, publicKey...)

	return sigScript
}

func CreateCoinbaseTransaction(crypt Cryptographic, txFees int) (repository.Transaction, int) {
	// First create the transaction with TxIns and TxOuts - tx id and txIn signature are not included yet
	txIns := make([]repository.TxIn, 0)
	txOuts := make([]repository.TxO, 0)

	txOut := repository.TxO{
		Value:        COINBASE_TRANSACTION_AMOUNT + txFees,
		ScriptPubKey: crypt.FirstcoinAddress,
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
	utils.PanicError(err)

	return msgHash.Sum(nil)
}

type prettyTxO struct {
	Address string
	Amount  int
}

func AreValidTransactions(txs []repository.Transaction) error {
	if len(txs) == 0 {
		return fmt.Errorf("Invalid transactions. Cant have empty transactions")
	}

	// first transaction in the list is always the coinbase transaction
	coinbaseTransaction := txs[0]

	if err := IsValidCoinbaseTransaction(coinbaseTransaction, txs[1:]); err != nil {
		return err
	}

	for _, transaction := range txs[1:] {
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

func CalculateTotalTxFees(txPool []repository.Transaction) (int, []repository.Transaction) {
	totalFees := 0
	uTxOSet := repository.GetEntireUTxOSet()
	txPoolToInclude := make([]repository.Transaction, 0)

	for _, tx := range txPool {
		totalInput, totalOutput := CalculateFeeForTx(tx, uTxOSet)

		if totalInput-totalOutput >= TRANSACTION_FEE {
			txPoolToInclude = append(txPoolToInclude, tx)
			totalFees += totalInput - totalOutput
		}
	}

	return totalFees, txPoolToInclude
}

func CalculateFeeForTx(tx repository.Transaction, uTxOSet repository.UTxOSetType) (int, int) {
	totalInput := 0
	totalOutput := 0

	for _, input := range tx.TxIns {
		tx := uTxOSet[repository.TxIDType(input.TxID)]
		inputAmount := tx.TxOuts[input.TxOIndex].Value
		totalInput += inputAmount
	}

	for _, output := range tx.TxOuts {
		totalOutput += output.Value
	}

	return totalInput, totalOutput
}

func IsValidCoinbaseTransaction(tx repository.Transaction, otherTxs []repository.Transaction) error {
	if len(tx.TxOuts) != 1 {
		return fmt.Errorf("Invalid coinbase transaction txOuts length > 0")
	}

	fees := tx.TxOuts[0].Value - COINBASE_TRANSACTION_AMOUNT
	totalFees, _ := CalculateTotalTxFees(otherTxs)

	if fees != totalFees {
		return fmt.Errorf("Invalid coinbase transaction. Fees are incorrect")
	}

	tID := GenerateTransactionID(tx)
	if !reflect.DeepEqual(tID, tx.ID) {
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
	spenderLedger := repository.GetUserLedger(w.Crypt.FirstcoinAddress)
	uTxOs := make([]TxIDIndexPair, 0)

	totalAmount := 0

	for _, tx := range spenderLedger {
		for index, uTxO := range tx.TxOuts {
			if totalAmount < amount+TRANSACTION_FEE && !isUTxOInTxPool(tx.ID) && uTxOBelongsToSpender(uTxO, w.Crypt.FirstcoinAddress) {
				uTxOs = append(uTxOs, TxIDIndexPair{
					TxID:     tx.ID,
					TxOIndex: index,
				})
				totalAmount += uTxO.Value
				break

			}

			if totalAmount >= amount+TRANSACTION_FEE {
				return uTxOs, totalAmount, nil
			}
		}
	}

	if totalAmount < amount {
		return nil, totalAmount, fmt.Errorf("insufficient funds or no available uTxOs")
	}

	if totalAmount < amount+TRANSACTION_FEE {
		return nil, totalAmount, fmt.Errorf("insufficient funds to include the tx fee of 1 coin")
	}

	return uTxOs, totalAmount, nil
}

func uTxOBelongsToSpender(uTxO repository.TxO, spenderPublicKey []byte) bool {
	return reflect.DeepEqual(uTxO.ScriptPubKey, spenderPublicKey)
}

// can only send to one receiver, and can get change. Bitcoin protocol allows for multiple receivers and senders in one tx. Potential TODO.
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
	if totalAmount > amount+TRANSACTION_FEE {
		change = totalAmount - (amount + TRANSACTION_FEE)
		changeTxO := repository.TxO{
			ScriptPubKey: w.Crypt.FirstcoinAddress,
			Value:        change,
		}

		txOs = append(txOs, changeTxO)
	}

	return txOs, nil
}

func (w *Wallet) validateTxInsCanServiceAmount(txIns []repository.TxIn, amount int) (bool, int) {
	totalAmount := 0

	for _, txIn := range txIns {
		spenderTxs := repository.GetUserLedger(w.Crypt.FirstcoinAddress)
		tx := spenderTxs[repository.TxIDType(txIn.TxID)]
		totalAmount += tx.TxOuts[txIn.TxOIndex].Value
	}

	return totalAmount >= amount+TRANSACTION_FEE, totalAmount
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

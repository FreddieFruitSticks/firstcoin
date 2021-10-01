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

	uTxOs, _, err := w.FindUTxOs(amount)
	if err != nil {
		return nil, 0, err
	}

	for _, uTxO := range uTxOs {
		txIn := repository.TxIn{
			UTxOID: repository.UTxOID{
				Address: uTxO.ID.Address,
				TxID:    uTxO.ID.TxID,
			},
			UTxOIndex: uTxO.Index,
		}
		txIns = append(txIns, txIn)
	}

	txOs, _ := w.GetTxOs(amount, receiverAddress, uTxOs)
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
		transaction.TxIns[i].Signature = signature
	}

	return &transaction, now, nil
}

func CreateCoinbaseTransaction(crypt Cryptographic, blockIndex int) (repository.Transaction, int) {
	// First create the transaction with TxIns and TxOuts - tx id and txIn signature are not included yet
	txIns := make([]repository.TxIn, 0)
	txOuts := make([]repository.TxO, 0)
	txIn := repository.TxIn{
		UTxOID: repository.UTxOID{
			Address: []byte{},
			TxID:    []byte{},
		},
		UTxOIndex: blockIndex,
	}

	txIns = append(txIns, txIn)

	txOut := repository.TxO{
		Amount:  COINBASE_TRANSACTION_AMOUNT,
		Address: crypt.PublicKey,
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

	// tx input signature is the tx id signed by the spender of coins - this will be used later to verify that the uTxO public key
	// is the true owner and authoriser of this signed transaction
	txInSignature := crypt.GenerateSignature(txID)
	txIns[0].Signature = txInSignature

	return transaction, now
}

// This is a SHA of all txIns (excluding signature - that gets added later) and txOuts
func GenerateTransactionID(transaction repository.Transaction) []byte {
	msgHash := sha256.New()
	concatTxIn := ""
	concatTxOut := ""

	for _, txIn := range transaction.TxIns {
		concatTxIn += string(txIn.UTxOID.Address) + strconv.Itoa(txIn.UTxOIndex)
	}

	for _, txOut := range transaction.TxOuts {
		concatTxOut += string(txOut.Address) + strconv.Itoa(txOut.Amount)
	}

	_, err := msgHash.Write([]byte(fmt.Sprintf("%s%s%d", concatTxIn, concatTxOut, transaction.Timestamp)))
	utils.CheckError(err)

	return msgHash.Sum(nil)
}

type prettyTxO struct {
	Address string
	Amount  int
}

func AreValidTransactions(transactions []repository.Transaction, blockIndex int) error {
	if len(transactions) == 0 {
		return fmt.Errorf("Invalid transactions. Cant have empty transactions")
	}

	// first transaction in the list is always the coinbase transaction
	coinbaseTransaction := transactions[0]

	if err := IsValidCoinbaseTransaction(coinbaseTransaction, blockIndex); err != nil {
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

	if err := AreValidTxIns(tx.TxIns, uTxOSet); err != nil {
		return fmt.Errorf("Invalid transaction, invalid txIn %+v", err)
	}

	if len(tx.TxOuts) < 1 {
		return fmt.Errorf("Invalid transaction: txOuts length must be greater than 0")
	}

	if err := AreValidTxOuts(tx.TxOuts); err != nil {
		return fmt.Errorf("Invalid transaction, invalid txOut %+v", err)
	}

	tID := GenerateTransactionID(tx)
	if !reflect.DeepEqual(tID, tx.ID) {
		return fmt.Errorf("Invalid transaction id: %s. generate: %s", tx.ID, tID)
	}

	for _, txIn := range tx.TxIns {
		if err := VerifySignature(txIn.Signature, txIn.UTxOID.Address, tx.ID); err != nil {
			return fmt.Errorf("Invalid transaction - signature verification failed: %+v", err.Error())

		}
	}
	if err := VerifyTransactionAmountCopy(tx, uTxOSet); err != nil {
		return fmt.Errorf("Invalid transaction - amount verification failed: %+v", err.Error())
	}

	return nil
}

func IsValidCoinbaseTransaction(transaction repository.Transaction, blockIndex int) error {
	if len(transaction.TxIns) != 1 {
		return fmt.Errorf("Invalid coinbase transaction txIns length > 0")
	}

	if len(transaction.TxOuts) != 1 {
		return fmt.Errorf("Invalid coinbase transaction txOuts length > 0")
	}

	if transaction.TxOuts[0].Amount != COINBASE_TRANSACTION_AMOUNT {
		return fmt.Errorf("Invalid coinbase transaction amount != COINBASE_TRANSACTION_AMOUNT")
	}

	if transaction.TxIns[0].UTxOIndex != blockIndex {
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
		uTxOId := txIn.UTxOID

		spenderLedger := uTxOSet[repository.PublicKeyAddressType(uTxOId.Address)]

		if len(spenderLedger) == 0 {
			return fmt.Errorf("spender does not exist in public ledger")
		}

		spenderUTxO := spenderLedger[repository.TxIDType(uTxOId.TxID)]
		if err := IsValidUTxOStructure(spenderUTxO); err != nil {
			return err
		}
		totalAmountFromUTxOs += spenderUTxO.Amount
	}

	if tx.TxOuts[0].Amount > totalAmountFromUTxOs {
		return fmt.Errorf("unspent transaction output does not have enough coin")
	}

	return nil
}

func IsValidUTxOStructure(uTxO repository.UTxO) error {
	if uTxO.Amount <= 0 {
		return fmt.Errorf("spender must reference a uTxO with more than 0 coins")
	}

	if len(uTxO.ID.Address) == 0 {
		return fmt.Errorf("uTxO address cannot be empty")
	}

	if len(uTxO.ID.TxID) == 0 {
		return fmt.Errorf("uTxO ID invalid")
	}

	if uTxO.Index < 0 {
		return fmt.Errorf("uTxO Index invalid - must be a valid block index")
	}

	return nil
}

func IsValidTxIn(txIn repository.TxIn, uTxOSet repository.UTxOSetType) error {
	spenderLedger, ok := uTxOSet[repository.PublicKeyAddressType(txIn.UTxOID.Address)]

	// This means that spender cant use the same uTxO in the same mempool. when validating entire pool the utxo will disappear.
	if !ok {
		return fmt.Errorf("invalid txIn. spender ledger does not exist")
	}

	_, ok = spenderLedger[repository.TxIDType(txIn.UTxOID.TxID)]
	if !ok {
		return fmt.Errorf("invalid txIn. spender uTxO does not exist")
	}

	if len(txIn.UTxOID.Address) == 0 {
		return fmt.Errorf("txIn UTxO address cannot be empty")
	}

	if len(txIn.UTxOID.TxID) == 0 {
		return fmt.Errorf("txIn UTxOID TxID cannot be empty")
	}

	if len(txIn.Signature) == 0 {
		return fmt.Errorf("txIn Signature cannot be empty")
	}

	return nil
}

func AreValidTxIns(txIns []repository.TxIn, uTxOSet repository.UTxOSetType) error {
	for index, txIn := range txIns {
		if err := IsValidTxIn(txIn, uTxOSet); err != nil {
			return fmt.Errorf("error in txIn number %d. error: %+v", index, err)
		}
	}

	return nil
}

func IsValidTxOutStructure(txOut repository.TxO) error {
	if len(txOut.Address) == 0 {
		return fmt.Errorf("invalid txOut: address must be valid public key")
	}

	if txOut.Amount <= 0 {
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

// finding the senders UTxOs that can service the Tx amount - currently the strategy is simply to take the first set of uTxOs
// that is not already included in the txPool. With the idea of parent-child txs can maybe have multiple txs on single uTxO???
func (w *Wallet) FindUTxOs(amount int) ([]repository.UTxO, int, error) {
	spenderUTxOs := repository.GetUserLedger(w.Crypt.PublicKey)
	uTxOs := make([]repository.UTxO, 0)

	totalAmount := 0

	for _, uTxO := range spenderUTxOs {
		if totalAmount < amount && !isUTxOInTxPool(uTxO) {
			uTxOs = append(uTxOs, uTxO)
			totalAmount += uTxO.Amount
			continue
		}

		if totalAmount >= amount {
			break
		}
	}

	if totalAmount < amount {
		return nil, totalAmount, fmt.Errorf("insufficient funds or no available uTxOs")
	}

	return uTxOs, totalAmount, nil
}

// can only send to one receiver, and can get change
func (w *Wallet) GetTxOs(amount int, receiverAddress []byte, uTxOs []repository.UTxO) ([]repository.TxO, error) {
	txOs := make([]repository.TxO, 0)

	valid, totalAmount := validateUTxOsCanServiceAmount(uTxOs, amount)
	if !valid {
		return nil, fmt.Errorf("uTxOs cannot service amount. uTxO total: %d. amount: %d", totalAmount, amount)
	}

	txO := repository.TxO{
		Address: receiverAddress,
		Amount:  amount,
	}

	txOs = append(txOs, txO)

	change := 0
	// deduct the difference between the total amount and the amount required, and add that as a repository.TxO to go back to the spender (as change)
	if totalAmount > amount {
		change = totalAmount - amount
		changeTxO := repository.TxO{
			Address: w.Crypt.PublicKey,
			Amount:  change,
		}

		txOs = append(txOs, changeTxO)
	}

	return txOs, nil
}

func validateUTxOsCanServiceAmount(uTxOs []repository.UTxO, amount int) (bool, int) {
	totalAmount := 0

	for _, uTxO := range uTxOs {
		totalAmount += uTxO.Amount
	}

	return totalAmount >= amount, totalAmount
}

func isUTxOInTxPool(uTxO repository.UTxO) bool {
	txPool := repository.GetTxPool()

	for _, v := range txPool {
		if reflect.DeepEqual(uTxO.ID.TxID, v.TxIns[0].UTxOID.TxID) {
			return true
		}
	}

	return false
}

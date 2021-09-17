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

type PublicKeyAddressType string
type TxIDType string
type UTxOSetType map[PublicKeyAddressType]map[TxIDType]UTxO

// to reference the UTxO we need the address and txID because the UTxO set is a map of maps
type UTxOID struct {
	Address []byte
	TxID    []byte
}

// transcation input refers to the giver of coins. Signature is signed with giver's private key
type TxIn struct {
	UTxOID    UTxOID
	UTxOIndex int // index is the block number or block height - this is to prevent duplicate signatures for exact same txs
	Signature []byte
}

// transaction ouput refers to the receiver of coins. Address is receiver's public key
type TxOut struct {
	Address []byte
	Amount  int
}

// unspend transaction outputs are the final transaction outs allocated to each receiver - it's ID is the ID of the transaction that created it
type UTxO struct {
	ID     UTxOID //ID is unique because of timestamp
	Index  int
	Amount int
}

type Wallet struct {
	UTxOSet UTxOSetType
	Crypt   Cryptographic
}

func NewWallet(u UTxOSetType, c Cryptographic) *Wallet {
	return &Wallet{
		UTxOSet: u,
		Crypt:   c,
	}
}

func CreateTransaction(receiverAddress []byte, amount int, uTxO *UTxO, crypt *Cryptographic) (Transaction, int) {
	txIns := make([]TxIn, 0)
	txOuts := make([]TxOut, 0)

	txIn := TxIn{
		UTxOID: UTxOID{
			Address: uTxO.ID.Address,
			TxID:    uTxO.ID.TxID,
		},
		UTxOIndex: uTxO.Index,
	}
	txIns = append(txIns, txIn)

	txOut := TxOut{
		Amount:  amount,
		Address: receiverAddress,
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
	signature := crypt.GenerateSignature(txID)
	txIn.Signature = signature

	return Transaction{
		ID:        txID,
		TxIns:     txIns,
		TxOuts:    txOuts,
		Timestamp: now,
	}, now
}

func CreateCoinbaseTransaction(crypt Cryptographic, blockIndex int) (Transaction, int) {
	// First create the transaction with TxIns and TxOuts - tx id and txIn signature are not included yet
	txIns := make([]TxIn, 0)
	txOuts := make([]TxOut, 0)
	txIn := TxIn{
		UTxOID: UTxOID{
			Address: []byte{},
			TxID:    []byte{},
		},
		UTxOIndex: blockIndex,
	}

	txIns = append(txIns, txIn)

	txOut := TxOut{
		Amount:  COINBASE_TRANSACTION_AMOUNT,
		Address: crypt.PublicKey,
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
	txInSignature := crypt.GenerateSignature(txID)
	txIns[0].Signature = txInSignature

	return transaction, now
}

// This is a SHA of all txIns (excluding signature - that gets added later) and txOuts
func GenerateTransactionID(transaction Transaction) []byte {
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

func (w *Wallet) IsValidTransaction(transaction Transaction) error {
	if len(transaction.TxIns) < 1 {
		return fmt.Errorf("Invalid transaction: txIns length must be > 0")
	}

	if err := AreValidTxIns(transaction.TxIns); err != nil {
		return fmt.Errorf("Invalid transaction, invalid txIn %+v", err)
	}

	if len(transaction.TxOuts) < 1 {
		return fmt.Errorf("Invalid transaction: txOuts length must be > 0")
	}

	if err := AreValidTxOuts(transaction.TxOuts); err != nil {
		return fmt.Errorf("Invalid transaction, invalid txOut %+v", err)
	}

	tID := GenerateTransactionID(transaction)
	if !reflect.DeepEqual(tID, transaction.ID) {
		return fmt.Errorf("Invalid transaction id")
	}

	if err := w.Crypt.VerifySignature(transaction.TxIns[0].Signature, transaction.TxIns[0].UTxOID.Address, transaction.ID); err != nil {
		return fmt.Errorf("Invalid transaction - signature verification failed: %+v", err.Error())

	}
	if err := w.VerifyTransactionAmount(transaction); err != nil {
		return fmt.Errorf("Invalid transaction - amount verification failed: %+v", err.Error())
	}

	return nil
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

func (w *Wallet) VerifyTransactionAmount(tx Transaction) error {
	if len(tx.TxIns) < 1 {
		return fmt.Errorf("Invalid transaction: txIns length must be > 0")
	}

	uTxOId := tx.TxIns[0].UTxOID

	spenderLedger := w.UTxOSet[PublicKeyAddressType(uTxOId.Address)]
	if len(spenderLedger) == 0 {
		return fmt.Errorf("spender does not exist in public ledger")
	}

	spenderUTxO := spenderLedger[TxIDType(uTxOId.TxID)]

	if err := IsValidUTxOStructure(spenderUTxO); err != nil {
		return err
	}

	if tx.TxOuts[0].Amount > spenderUTxO.Amount {
		return fmt.Errorf("unspent transaction output does not have enough coin")
	}

	return nil
}

func IsValidUTxOStructure(uTxO UTxO) error {
	if uTxO.Amount <= 0 {
		return fmt.Errorf("spender must spend more than 0 coins")
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

func IsValidTxInStructure(txIn TxIn) error {
	if len(txIn.UTxOID.Address) == 0 {
		return fmt.Errorf("txIn UTxO address cannot be empty")
	}

	if len(txIn.UTxOID.TxID) == 0 {
		return fmt.Errorf("txIn UTxOID TxID cannot be empty")
	}

	if txIn.UTxOIndex == 0 {
		return fmt.Errorf("txIn UTxOIndex cannot be 0")
	}

	if len(txIn.Signature) == 0 {
		return fmt.Errorf("txIn Signature cannot be empty")
	}

	return nil
}

func AreValidTxIns(txIns []TxIn) error {
	for index, txIn := range txIns {
		if err := IsValidTxInStructure(txIn); err != nil {
			return fmt.Errorf("error in txIn number %d. error: %+v", index, err)
		}
	}

	return nil
}

func IsValidTxOutStructure(txOut TxOut) error {
	if len(txOut.Address) == 0 {
		return fmt.Errorf("invalid txOut: address must be valid public key")
	}

	if txOut.Amount <= 0 {
		return fmt.Errorf("invalid txOut: amount must be > 0")
	}

	return nil
}

func AreValidTxOuts(txOuts []TxOut) error {
	for index, txOut := range txOuts {
		if err := IsValidTxOutStructure(txOut); err != nil {
			return fmt.Errorf("invalid txOut number %d. error: %+v", index, err)
		}
	}

	return nil
}

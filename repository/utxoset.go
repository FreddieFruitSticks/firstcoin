package repository

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
)

var uTxOSet = UTxOSetType(make(map[TxIDType]Transaction))

type TxIDType string
type UTxOSetType map[TxIDType]Transaction
type UserWalletType map[TxIDType]Transaction // wallet is basically the subset of UTxOSet that concerns the user

// transcation input refers to the giver of coins. Signature is signed with giver's private key
type TxIn struct {
	TxID            []byte `json:"txid"`
	TxOIndex        int    `json:"vout"`
	ScriptSignature []byte `json:"scriptSig"`
}

// transaction ouput refers to the receiver of coins. Address is receiver's public key
type TxO struct {
	ScriptPubKey []byte `json:"scriptPubKey"`
	Value        int    `json:"value"`
}

func GetEntireUTxOSet() UTxOSetType {
	return uTxOSet
}

func GetUserLedger(publicKey []byte) UserWalletType {
	wallet := make(map[TxIDType]Transaction)

	for _, tx := range uTxOSet {
		for _, txO := range tx.TxOuts {
			if reflect.DeepEqual(txO.ScriptPubKey, publicKey) {
				wallet[TxIDType(tx.ID)] = tx
				break
			}
		}
	}

	return wallet
}

func AddTxToUTxOSet(tx Transaction) {
	AddTxSpecifiedToUTxOSet(tx, uTxOSet)
}

func AddTxToUTxOSetCopy(tx Transaction, uTxOSetCopy UTxOSetType) {
	AddTxSpecifiedToUTxOSet(tx, uTxOSetCopy)
}

func AddTxSpecifiedToUTxOSet(tx Transaction, uTxOSet UTxOSetType) {
	uTxOSet[TxIDType(tx.ID)] = tx
}

func RemoveTxOFromUTxOSet(txID TxIDType, txIn TxIn) {
	RemoveTxOFromUTxOCopy(txID, txIn, uTxOSet)
}

func RemoveTxOFromUTxOCopy(txID TxIDType, txIn TxIn, uTxOSet UTxOSetType) {
	index := txIn.TxOIndex
	uTxOID := txIn.TxID
	txOs := uTxOSet[TxIDType(uTxOID)].TxOuts

	remainingTxOs := append(txOs[:index], txOs[index+1:]...)
	if len(remainingTxOs) == 0 {
		delete(uTxOSet, TxIDType(uTxOID))
	} else {
		tx := uTxOSet[TxIDType(uTxOID)]
		tx.TxOuts = remainingTxOs
		uTxOSet[TxIDType(uTxOID)] = tx
	}
}

func ClearUTxOSet() {
	uTxOSet = UTxOSetType(make(map[TxIDType]Transaction))
}

func (t TxIn) String() string {
	return fmt.Sprintf("{\nTxID: %s\nuTxOIndex: %+v\nscriptSig: %+v\n}\n", t.TxID, t.TxOIndex, Base64Encode(t.ScriptSignature))
}

func (t TxO) String() string {
	return fmt.Sprintf("{\nAddress: %s\nAmount: %+v\n}\n", string(t.ScriptPubKey), t.Value)
}

func Base64Encode(message []byte) []byte {
	b := make([]byte, base64.StdEncoding.EncodedLen(len(message)))
	base64.StdEncoding.Encode(b, message)
	return b
}

func CopyUTxOSet() UTxOSetType {
	uTxOSetCopy := make(map[TxIDType]Transaction)
	for txID, tx := range uTxOSet {
		txCopy := Transaction{}
		bytes, _ := json.Marshal(tx)
		json.Unmarshal(bytes, &txCopy)
		uTxOSetCopy[txID] = txCopy
	}

	return uTxOSetCopy
}

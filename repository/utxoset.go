package repository

import (
	"encoding/base64"
	"fmt"
)

var uTxOSet = UTxOSetType(make(map[PublicKeyAddressType]map[TxIDType]UTxO))

type PublicKeyAddressType string
type TxIDType string
type UTxOSetType map[PublicKeyAddressType]map[TxIDType]UTxO
type userWalletType map[TxIDType]UTxO

// to reference the UTxO we need the address and txID because the UTxO set is a map of maps
type UTxOID struct {
	Address []byte `json:"Address"`
	TxID    []byte `json:"TxID"`
}

// transcation input refers to the giver of coins. Signature is signed with giver's private key
type TxIn struct {
	UTxOID    UTxOID
	UTxOIndex int // index is the block number or block height - this is to prevent duplicate signatures for exact same txs
	Signature []byte
}

// transaction ouput refers to the receiver of coins. Address is receiver's public key
type TxO struct {
	Address []byte `json:"Address"`
	Amount  int    `json:"Amount"`
}

// unspend transaction outputs are the final transaction outs allocated to each receiver - it's ID is the ID of the transaction that created it
type UTxO struct {
	ID     UTxOID //ID is unique because of timestamp
	Index  int
	Amount int
}

func GetEntireUTxOSet() UTxOSetType {
	return uTxOSet
}

func GetUserLedger(publicKey []byte) userWalletType {
	return uTxOSet[PublicKeyAddressType(publicKey)]
}

func AddTxOToReceiver(txId []byte, blockIndex int, txO TxO) {
	AddTxOToReceiverSet(txId, blockIndex, txO, uTxOSet)
}

func AddTxOToReceiverCopy(txId []byte, blockIndex int, txO TxO, uTxOSetCopy UTxOSetType) {
	AddTxOToReceiverSet(txId, blockIndex, txO, uTxOSetCopy)
}

func AddTxOToReceiverSet(txId []byte, blockIndex int, txO TxO, uTxOSet UTxOSetType) {
	uTxO := UTxO{
		ID: UTxOID{
			Address: txO.Address,
			TxID:    txId,
		},
		Amount: txO.Amount,
		Index:  blockIndex,
	}

	receiverUTxOs := uTxOSet[PublicKeyAddressType(txO.Address)]
	if receiverUTxOs == nil {
		uTxOMap := make(map[TxIDType]UTxO)
		uTxOMap[TxIDType(txId)] = uTxO

		uTxOSet[PublicKeyAddressType(txO.Address)] = uTxOMap
		return
	}

	uTxOSet[PublicKeyAddressType(txO.Address)][TxIDType(txId)] = uTxO
}

func RemoveUTxOFromSender(txIn TxIn) {
	RemoveUTxOFromSenderCopy(txIn, uTxOSet)
}

func RemoveUTxOFromSenderCopy(txIn TxIn, uTxOSet UTxOSetType) {
	delete(uTxOSet[PublicKeyAddressType(txIn.UTxOID.Address)], TxIDType(txIn.UTxOID.TxID))
}

func ClearUTxOSet() {
	uTxOSet = UTxOSetType(make(map[PublicKeyAddressType]map[TxIDType]UTxO, 0))
}

func (t TxIn) String() string {
	return fmt.Sprintf("{\nuTxOID: %s\nuTxOIndex: %+v\nUtxOuts: %+v\n}\n", t.UTxOID, t.UTxOIndex, Base64Encode(t.Signature))
}

func (t TxO) String() string {
	return fmt.Sprintf("{\nAddress: %s\nAmount: %+v\n}\n", Base64Encode(t.Address), t.Amount)
}

func (t UTxOID) String() string {
	return fmt.Sprintf("{\nAddress: %s\nTxID: %+v\n}\n", Base64Encode(t.Address), t.TxID)
}

func (t UTxO) String() string {
	return fmt.Sprintf("{\nId: %s\nIndex: %+v\nAmount: %+v\n}\n", t.ID, t.Index, t.Amount)
}

func Base64Encode(message []byte) []byte {
	b := make([]byte, base64.StdEncoding.EncodedLen(len(message)))
	base64.StdEncoding.Encode(b, message)
	return b
}

func CopyUTxOSet() UTxOSetType {
	uTxOSetCopy := make(map[PublicKeyAddressType]map[TxIDType]UTxO)

	for k, v := range uTxOSet {
		id := make(map[TxIDType]UTxO)
		for k1, v1 := range v {
			id[k1] = v1
		}
		uTxOSetCopy[k] = id
	}

	return uTxOSetCopy
}

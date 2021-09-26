package repository

var uTxOSet = UTxOSetType(make(map[PublicKeyAddressType]map[TxIDType]UTxO, 0))

type PublicKeyAddressType string
type TxIDType string
type UTxOSetType map[PublicKeyAddressType]map[TxIDType]UTxO
type userWalletType map[TxIDType]UTxO

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
type TxO struct {
	Address []byte
	Amount  int
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
	delete(uTxOSet[PublicKeyAddressType(txIn.UTxOID.Address)], TxIDType(txIn.UTxOID.TxID))
}

func ClearUTxOSet() {
	uTxOSet = UTxOSetType(make(map[PublicKeyAddressType]map[TxIDType]UTxO, 0))
}

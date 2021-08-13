package coin

import (
	"blockchain/utils"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

const difficulty = "00000"

type Block struct {
	Index        int    `json:"index"`
	PreviousHash []byte `json:"previousHash"`
	Data         string `json:"data"`
	Timestamp    int    `json:"timeStamp"`
	Difficulty   int    `json:"difficulty"`
	Nonce        int    `json:"nonce"`
	Hash         []byte `json:"hash"`
}

func calculateBlockHash(index int, previousHash []byte, timestamp int, data string) []byte {
	msgHash := sha256.New()
	_, err := msgHash.Write([]byte(fmt.Sprintf("%d%s%d%s", index, string(previousHash), timestamp, data)))
	utils.CheckError(err)

	return msgHash.Sum(nil)
}

func GenesisBlock() Block {
	beginning := int(time.Date(2021, time.August, 13, 0, 0, 0, 0, time.UTC).UnixNano())
	data := "Genesis Block"
	prevHash := []byte{}

	return Block{
		Index:        0,
		PreviousHash: prevHash,
		Data:         data,
		Timestamp:    beginning,
		Hash:         calculateBlockHash(0, prevHash, beginning, data),
	}
}

func (b *Block) IsValidBlock(previousBlock Block) bool {
	if previousBlock.Index+1 != b.Index {
		return false
	}

	if !reflect.DeepEqual(b.PreviousHash, previousBlock.Hash) {
		return false
	}

	if !reflect.DeepEqual(calculateBlockHash(b.Index, b.PreviousHash, b.Timestamp, b.Data), b.Hash) {
		return false
	}

	if !ValidateProofOfWork(b.Hash, b.Nonce) {
		return false
	}

	return true
}

func work(hash string) bool {
	return hash[0:len(difficulty)] == difficulty
}

func ProofOfWork(hash []byte) int {
	var nonce int

	s := fmt.Sprintf("%s%d", string(hash), nonce)
	h := hashBytes([]byte(s))

	for !work(h) {
		nonce++
		s := fmt.Sprintf("%s%d", string(hash), nonce)
		h = hashBytes([]byte(s))
	}

	return nonce
}

func ValidateProofOfWork(hash []byte, nonce int) bool {
	s := fmt.Sprintf("%s%d", string(hash), nonce)
	h := hashBytes([]byte(s))

	return work(h)
}

func Hash(block Block) string {
	blockString, err := json.Marshal(block)
	utils.CheckError(err)

	sha1Hash := hashBytes(blockString)

	return sha1Hash
}

func hashBytes(blockString []byte) string {
	hash := sha256.New()
	hash.Write(blockString)
	sha1Hash := hex.EncodeToString(hash.Sum(nil))
	return sha1Hash
}

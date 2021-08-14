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

const (
	difficultyLevel = 5
	GensisBlockData = "Genesis Block"
)

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
	var prevHash []byte
	beginning := int(time.Date(2021, time.August, 13, 0, 0, 0, 0, time.UTC).UnixNano())

	return Block{
		Index:        0,
		PreviousHash: prevHash,
		Data:         GensisBlockData,
		Timestamp:    beginning,
		Hash:         calculateBlockHash(0, prevHash, beginning, GensisBlockData),
	}
}

func (b *Block) IsGenesisBlock() bool {
	beginning := int(time.Date(2021, time.August, 13, 0, 0, 0, 0, time.UTC).UnixNano())
	return b.Index == 0 && len(b.PreviousHash) == 0 && b.Data == GensisBlockData && b.Timestamp == beginning
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

	if !validateNewBlockDifficulty(*b) {
		return false
	}

	return true
}

func validateNewBlockDifficulty(b Block) bool {
	s := fmt.Sprintf("%s%d", string(b.Hash), b.Nonce)
	powHash := hashBytes([]byte(s))
	difficulty := generateDifficulty()

	return validateDifficulty(powHash, difficulty)
}

func validateDifficulty(powHash string, difficulty []byte) bool {
	return powHash[0:difficultyLevel] == string(difficulty)
}

func generateDifficulty() []byte {
	var difficulty []byte
	for i := 0; i < difficultyLevel; i++ {

		// 48 is the value of 0 string
		difficulty = append(difficulty, 48)
	}

	return difficulty
}

// Validating that the first number of chars of the powHash are 0's
func ProofOfWork(blockHash []byte) int {
	var nonce int

	s := fmt.Sprintf("%s%d", string(blockHash), nonce)
	powHash := hashBytes([]byte(s))
	difficulty := generateDifficulty()

	for !validateDifficulty(powHash, difficulty) {
		nonce++
		s := fmt.Sprintf("%s%d", string(blockHash), nonce)
		powHash = hashBytes([]byte(s))
	}

	return nonce
}

func ValidateProofOfWork(hash []byte, nonce int) bool {
	s := fmt.Sprintf("%s%d", string(hash), nonce)
	h := hashBytes([]byte(s))
	difficulty := generateDifficulty()

	return validateDifficulty(h, difficulty)
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

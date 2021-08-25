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
	GensisBlockData = "Genesis Block"
)

type Block struct {
	Index           int    `json:"index"`
	PreviousHash    []byte `json:"previousHash"`
	Data            string `json:"data"`
	Timestamp       int    `json:"timeStamp"`
	DifficultyLevel int    `json:"difficultyLevel"`
	Nonce           int    `json:"nonce"`
	Hash            []byte `json:"hash"`
}

func calculateBlockHash(index int, previousHash []byte, timestamp int, data string, difficultyLevel int) []byte {
	msgHash := sha256.New()
	_, err := msgHash.Write([]byte(fmt.Sprintf("%d%s%d%s%d", index, string(previousHash), timestamp, data, difficultyLevel)))
	utils.CheckError(err)

	return msgHash.Sum(nil)
}

func GenesisBlock(seedDifficultyLevel int) Block {
	var prevHash []byte
	beginning := int(time.Date(2021, time.August, 13, 0, 0, 0, 0, time.UTC).UnixNano())
	blockHash := calculateBlockHash(0, prevHash, beginning, GensisBlockData, seedDifficultyLevel)

	return Block{
		Index:           0,
		PreviousHash:    prevHash,
		Data:            GensisBlockData,
		Timestamp:       beginning,
		Hash:            blockHash,
		DifficultyLevel: seedDifficultyLevel,
		Nonce:           ProofOfWork(blockHash, seedDifficultyLevel),
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

	if !reflect.DeepEqual(calculateBlockHash(b.Index, b.PreviousHash, b.Timestamp, b.Data, b.DifficultyLevel), b.Hash) {
		return false
	}

	if !ValidateProofOfWork(b.Hash, b.Nonce, b.DifficultyLevel) {
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
	difficultyString := generateDifficulty(b.DifficultyLevel)

	return validateDifficulty(powHash, difficultyString, b.DifficultyLevel)
}

func validateDifficulty(powHash string, difficultyString []byte, difficultyLevel int) bool {
	return powHash[0:difficultyLevel] == string(difficultyString)
}

func generateDifficulty(difficultyLevel int) []byte {
	var difficulty []byte
	for i := 0; i < difficultyLevel; i++ {

		// 48 is the value of the 0 string
		difficulty = append(difficulty, 48)
	}

	return difficulty
}

// Validating that the first number of chars of the powHash are 0's
func ProofOfWork(blockHash []byte, difficultyLevel int) int {
	var nonce int

	s := fmt.Sprintf("%s%d", string(blockHash), nonce)
	powHash := hashBytes([]byte(s))
	difficulty := generateDifficulty(difficultyLevel)

	for !validateDifficulty(powHash, difficulty, difficultyLevel) {
		nonce++
		s := fmt.Sprintf("%s%d", string(blockHash), nonce)
		powHash = hashBytes([]byte(s))
	}

	return nonce
}

func ValidateProofOfWork(hash []byte, nonce int, difficultyLevel int) bool {
	s := fmt.Sprintf("%s%d", string(hash), nonce)
	h := hashBytes([]byte(s))
	difficulty := generateDifficulty(difficultyLevel)

	return validateDifficulty(h, difficulty, difficultyLevel)
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

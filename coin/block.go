package coin

import (
	"blockchain/repository"
	"blockchain/utils"
	"blockchain/wallet"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

type Block struct {
	Index           int                      `json:"index"`
	PreviousHash    []byte                   `json:"previousHash"`
	Transactions    []repository.Transaction `json:"transactions"`
	Timestamp       int                      `json:"timestamp"`
	DifficultyLevel int                      `json:"difficultyLevel"`
	Nonce           int                      `json:"nonce"`
	Hash            []byte                   `json:"hash"`
}

func calculateBlockHash(index int, previousHash []byte, timestamp int, transactions []repository.Transaction, difficultyLevel int) []byte {
	msgHash := sha256.New()

	// TODO: Does POW hash calculation contain transactions??
	concatenatedTransactionIDs := concatTransactionIDs(transactions)
	_, err := msgHash.Write([]byte(fmt.Sprintf("%d%s%d%s%d", index, string(previousHash), timestamp, concatenatedTransactionIDs, difficultyLevel)))
	utils.CheckError(err)

	return msgHash.Sum(nil)
}

func GenesisBlock(seedDifficultyLevel int, transactionPool []repository.Transaction) Block {
	var prevHash []byte
	beginning := int(time.Date(2021, time.August, 13, 0, 0, 0, 0, time.UTC).UnixNano())
	blockHash := calculateBlockHash(0, prevHash, beginning, transactionPool, seedDifficultyLevel)

	return Block{
		Index:           0,
		PreviousHash:    prevHash,
		Transactions:    transactionPool,
		Timestamp:       beginning,
		Hash:            blockHash,
		DifficultyLevel: seedDifficultyLevel,
		Nonce:           ProofOfWork(blockHash, seedDifficultyLevel),
	}
}

func (b *Block) IsGenesisBlock() error {
	beginning := int(time.Date(2021, time.August, 13, 0, 0, 0, 0, time.UTC).UnixNano())
	if b.Index != 0 {
		return fmt.Errorf("Genesis block must have 0 index")
	}

	if len(b.PreviousHash) != 0 {
		return fmt.Errorf("Genesis block must have 0 length previous hash")
	}

	if b.Timestamp != beginning {
		return fmt.Errorf("Genesis block must have timestamp set to beginning")
	}

	return nil
}

func (b *Block) IsValidBlock(previousBlock Block) error {
	if previousBlock.Index+1 != b.Index {
		return fmt.Errorf("Invalid block: %s", "invalid index")
	}

	if !reflect.DeepEqual(b.PreviousHash, previousBlock.Hash) {
		return fmt.Errorf("Invalid block: %s", "invalid previous block hash")
	}

	if !reflect.DeepEqual(calculateBlockHash(b.Index, b.PreviousHash, b.Timestamp, b.Transactions, b.DifficultyLevel), b.Hash) {
		return fmt.Errorf("Invalid block: %s", "invalid block hash")
	}

	if !ValidateProofOfWork(b.Hash, b.Nonce, b.DifficultyLevel) {
		return fmt.Errorf("Invalid block: %s", "invalid pow")
	}

	if !validateNewBlockDifficulty(*b) {
		return fmt.Errorf("Invalid block: %s", "invalid difficulty")
	}

	if b.Timestamp <= previousBlock.Timestamp {
		return fmt.Errorf("Invalid block: %s", "invalid timestamps")
	}

	if err := wallet.AreValidTransactions(b.Transactions, b.Index); err != nil {
		return fmt.Errorf("Invalid block: %s. error: %s", "invalid transactions", err.Error())
	}

	// validate that the current block's timestamp isnt more than 10s in the future - we allow a certain error in time registration
	// need to be careful with this value and time to mine a block
	if b.Timestamp > int(time.Now().UnixNano())+10*NANO_SECONDS {
		return fmt.Errorf("Invalid block: %s", "invalid block timestamp")
	}

	return nil
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

// a SHA version of a transaction is a concatenation of all transaction IDs and all transaction input signatures
func concatTransactionIDs(transactions []repository.Transaction) []byte {
	concatTransaction := []byte{}

	for _, transaction := range transactions {
		concatTxInSignatures := []byte{}
		for _, txIn := range transaction.TxIns {
			concatTxInSignatures = append(concatTxInSignatures, txIn.Signature...)
		}

		concatTransaction = append(concatTransaction, transaction.ID...)
		concatTransaction = append(concatTransaction, concatTxInSignatures...)
	}
	msgHash := sha256.New()
	_, err := msgHash.Write(concatTransaction)
	if err != nil {
		fmt.Printf("error hashing concatenated IDs %s", err.Error())
	}

	return msgHash.Sum(nil)
}

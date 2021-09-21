package coin

import (
	"blockchain/wallet"
	"fmt"
	"time"
)

const (
	BLOCK_GENERATION_INTERVAL      = 10         //seconds
	DIFFICULTY_ADJUSTMENT_INTERVAL = 10         //seconds
	NANO_SECONDS                   = 1000000000 //number of nanoseconds in 1 second
)

type Blockchain struct {
	Blocks []Block `json:"blocks"`
}

func NewBlockchain(b []Block) *Blockchain {
	return &Blockchain{
		Blocks: b,
	}
}

func (b *Blockchain) AddBlock(bl Block) {
	b.Blocks = append(b.Blocks, bl)
}

func (b *Blockchain) GetLastBlock() Block {
	return b.Blocks[len(b.Blocks)-1]
}

func (b *Blockchain) SetBlockchain(blocks []Block) {
	b.Blocks = blocks
}

func (b *Blockchain) GenerateNextBlock(transactionPool *[]wallet.Transaction) Block {
	previousBlock := b.GetLastBlock()
	now := int(time.Now().UnixNano())
	currentDifficultyLevel := b.getDifficultyLevel()
	fmt.Println("current difficulty: ", currentDifficultyLevel)
	hash := calculateBlockHash(previousBlock.Index+1, previousBlock.Hash, now, *transactionPool, currentDifficultyLevel)
	nonce := ProofOfWork(hash, currentDifficultyLevel)

	return Block{
		Index:           previousBlock.Index + 1,
		PreviousHash:    previousBlock.Hash,
		Transactions:    *transactionPool,
		Timestamp:       now,
		Hash:            hash,
		Nonce:           nonce,
		DifficultyLevel: currentDifficultyLevel,
	}
}

// Always favour the chain with the most work - it is sufficient to check the DifficultyLevel attribute on the block because this is validated in the IsValidBlock method
func (b *Blockchain) ReplaceBlockchain(bc Blockchain) {
	if bc.cumulativeDifficulty() > b.cumulativeDifficulty() {
		b.SetBlockchain(bc.Blocks)
	}
}

// calculate the difficulty of the block chain
func (b *Blockchain) cumulativeDifficulty() int {
	cumulativeDifficulty := 0
	for _, block := range b.Blocks {
		cumulativeDifficulty += block.DifficultyLevel
	}

	return cumulativeDifficulty
}

// Difficulty level is decreased by 1 if time between last 10 blocks > 200s (twice DIFFICULTY_ADJUSTMENT_INTERVAL * BLOCK_GENERATION_INTERVAL),
// and increased by 1 if time < 50s (half DIFFICULTY_ADJUSTMENT_INTERVAL * BLOCK_GENERATION_INTERVAL). This keeps it roughly 100s for 10 blocks
func (b *Blockchain) getDifficultyLevel() int {
	if b.GetLastBlock().Index%DIFFICULTY_ADJUSTMENT_INTERVAL == 0 && b.GetLastBlock().Index != 0 {
		fmt.Println((b.GetLastBlock().Timestamp - b.Blocks[len(b.Blocks)-DIFFICULTY_ADJUSTMENT_INTERVAL].Timestamp) / NANO_SECONDS)
		if (b.GetLastBlock().Timestamp - b.Blocks[len(b.Blocks)-DIFFICULTY_ADJUSTMENT_INTERVAL].Timestamp) >= 2*DIFFICULTY_ADJUSTMENT_INTERVAL*BLOCK_GENERATION_INTERVAL*NANO_SECONDS {
			return b.GetLastBlock().DifficultyLevel - 1
		}
		if (b.GetLastBlock().Timestamp - b.Blocks[len(b.Blocks)-DIFFICULTY_ADJUSTMENT_INTERVAL].Timestamp) <= 0.5*DIFFICULTY_ADJUSTMENT_INTERVAL*BLOCK_GENERATION_INTERVAL*NANO_SECONDS {
			return b.GetLastBlock().DifficultyLevel + 1
		}
	}
	return b.GetLastBlock().DifficultyLevel
}

func (b *Blockchain) IsValidBlockchain(u *wallet.UTxOSetType) bool {
	if !b.Blocks[0].IsGenesisBlock() {
		return false
	}

	for i := 1; i < len(b.Blocks); i++ {
		if !b.Blocks[i].IsValidBlock(b.Blocks[i-1], u) {
			return false
		}
	}

	return true
}

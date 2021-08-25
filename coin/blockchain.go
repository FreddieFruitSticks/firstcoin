package coin

import (
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

func NewBlockchain(b []Block, d int) *Blockchain {
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

func (b *Blockchain) GenerateNextBlock(blockData string) Block {
	previousBlock := b.GetLastBlock()
	now := int(time.Now().UnixNano())
	currentDifficultyLevel := b.getDifficultyLevel()
	fmt.Println("current difficulty: ", currentDifficultyLevel)
	hash := calculateBlockHash(previousBlock.Index+1, previousBlock.Hash, now, blockData, currentDifficultyLevel)
	nonce := ProofOfWork(hash, currentDifficultyLevel)

	return Block{
		Index:           previousBlock.Index + 1,
		PreviousHash:    previousBlock.Hash,
		Data:            blockData,
		Timestamp:       now,
		Hash:            hash,
		Nonce:           nonce,
		DifficultyLevel: currentDifficultyLevel,
	}
}

// Always favour the longest chain - most work
func (b *Blockchain) ReplaceBlockchain(bc Blockchain) {
	if bc.IsValidBlockchain() && len(bc.Blocks) > len(b.Blocks) {
		b.SetBlockchain(bc.Blocks)
	}
}

// Difficulty level is decreased by 1 if time between last 10 blocks > 200s, and increased by 1 if time < 50s. This keeps it roughly 100s for 10 blocks
func (b *Blockchain) getDifficultyLevel() int {
	if b.GetLastBlock().Index%DIFFICULTY_ADJUSTMENT_INTERVAL == 0 && b.GetLastBlock().Index != 0 {
		fmt.Println((b.GetLastBlock().Timestamp - b.Blocks[len(b.Blocks)-DIFFICULTY_ADJUSTMENT_INTERVAL].Timestamp) / NANO_SECONDS)
		if (b.GetLastBlock().Timestamp-b.Blocks[len(b.Blocks)-DIFFICULTY_ADJUSTMENT_INTERVAL].Timestamp)/NANO_SECONDS >= 2*DIFFICULTY_ADJUSTMENT_INTERVAL*BLOCK_GENERATION_INTERVAL {
			return b.GetLastBlock().DifficultyLevel - 1
		}
		if (b.GetLastBlock().Timestamp-b.Blocks[len(b.Blocks)-DIFFICULTY_ADJUSTMENT_INTERVAL].Timestamp)/NANO_SECONDS <= 0.5*DIFFICULTY_ADJUSTMENT_INTERVAL*BLOCK_GENERATION_INTERVAL {
			return b.GetLastBlock().DifficultyLevel + 1
		}
	}
	return b.GetLastBlock().DifficultyLevel
}

func (b *Blockchain) IsValidBlockchain() bool {
	if !b.Blocks[0].IsGenesisBlock() {
		return false
	}

	for i := 1; i < len(b.Blocks); i++ {
		if !b.Blocks[i].IsValidBlock(b.Blocks[i-1]) {
			return false
		}
	}

	return true
}

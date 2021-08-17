package coin

import (
	"time"
)

const (
	BLOCK_GENERATION_INTERVAL      = 10
	DIFFICULTY_ADJUSTMENT_INTERVAL = 10
)

type Blockchain struct {
	Blocks                 []Block `json:"blocks"`
	CurrentDifficultyLevel int     `json:"currentDifficultyLevel"`
}

func NewBlockchain(b []Block, d int) *Blockchain {
	return &Blockchain{
		Blocks:                 b,
		CurrentDifficultyLevel: d,
	}
}

func (b *Blockchain) AddBlock(bl Block) {
	b.Blocks = append(b.Blocks, bl)
}

func (b *Blockchain) GetLastBlock() Block {
	return b.Blocks[len(b.Blocks)-1]
}

func (b *Blockchain) SetBlockchain(blocks []Block, currentDifficultyLevel int) {
	b.Blocks = blocks
	b.CurrentDifficultyLevel = currentDifficultyLevel
}

func (b *Blockchain) GenerateNextBlock(blockData string) Block {
	previousBlock := b.GetLastBlock()
	now := int(time.Now().UnixNano())
	hash := calculateBlockHash(previousBlock.Index+1, previousBlock.Hash, now, blockData, b.CurrentDifficultyLevel)
	nonce := ProofOfWork(hash, b.CurrentDifficultyLevel)

	return Block{
		Index:           previousBlock.Index + 1,
		PreviousHash:    previousBlock.Hash,
		Data:            blockData,
		Timestamp:       now,
		Hash:            hash,
		Nonce:           nonce,
		DifficultyLevel: b.CurrentDifficultyLevel,
	}
}

func (b *Blockchain) ReplaceBlockchain(bc Blockchain) {
	if bc.IsValidBlockchain() && len(bc.Blocks) > len(b.Blocks) {
		b.SetBlockchain(bc.Blocks, bc.CurrentDifficultyLevel)
	}
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

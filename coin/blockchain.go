package coin

import (
	"reflect"
	"time"
)

type Blockchain struct {
	Blocks []Block `json:"blocks"`
}

func NewBlockchain() Blockchain {
	return Blockchain{
		Blocks: make([]Block, 0),
	}
}

func (b *Blockchain) AddBlock(bl Block) {
	b.Blocks = append(b.Blocks, bl)
}

func (b *Blockchain) GetLastBlock() Block {
	return b.Blocks[len(b.Blocks)-1]
}

func (b *Blockchain) GenerateNextBlock(blockData string) Block {
	previousBlock := b.GetLastBlock()
	now := int(time.Now().UnixNano())
	hash := calculateBlockHash(previousBlock.Index+1, previousBlock.Hash, now, blockData)
	nonce := ProofOfWork(hash)

	return Block{
		Index:        previousBlock.Index + 1,
		PreviousHash: previousBlock.Hash,
		Data:         blockData,
		Timestamp:    now,
		Hash:         hash,
		Nonce:        nonce,
	}
}

func (b *Blockchain) IsValidBlockchain() bool {
	if !reflect.DeepEqual(b.Blocks[0], GenesisBlock) {
		return false
	}

	for index, block := range b.Blocks[1:] {
		if !block.IsValidBlock(b.Blocks[index-1]) {
			return false
		}
	}

	return true
}

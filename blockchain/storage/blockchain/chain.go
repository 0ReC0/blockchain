package blockchain

import (
	"fmt"
	"sync"

	"blockchain/storage/txpool"
)

type Blockchain struct {
	Blocks []*Block
	mu     sync.Mutex
}

func NewBlockchain() *Blockchain {
	return &Blockchain{
		Blocks: []*Block{NewGenesisBlock()},
	}
}

func NewGenesisBlock() *Block {
	return NewBlock(0, "0", []*txpool.Transaction{}, "genesis")
}

func (bc *Blockchain) AddBlock(block *Block) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if bc.HasBlock(block.Hash) {
		return
	}

	bc.Blocks = append(bc.Blocks, block)
}

func (bc *Blockchain) GetBlockByNumber(blockNumber interface{}) *Block {
	// Предположим, что blockNumber — это строка вида "0x1" или число
	numStr, ok := blockNumber.(string)
	if !ok {
		return nil
	}

	// Упрощённый парсинг hex
	var num int64
	_, err := fmt.Sscanf(numStr, "0x%x", &num)
	if err != nil {
		return nil
	}

	for _, block := range bc.Blocks {
		if block.Index == num {
			return block
		}
	}
	return nil
}
func (bc *Blockchain) GetLatestBlock() *Block {
	if len(bc.Blocks) == 0 {
		return nil
	}
	return bc.Blocks[len(bc.Blocks)-1]
}
func (chain *Blockchain) HasBlock(hash string) bool {
	for _, b := range chain.Blocks {
		if b.Hash == hash {
			return true
		}
	}
	return false
}

package blockchain

import (
	"fmt"

	"blockchain/storage/txpool"
)

type Blockchain struct {
	Blocks []*Block
}

func NewBlockchain() *Blockchain {
	return &Blockchain{
		Blocks: []*Block{NewGenesisBlock()},
	}
}

func NewGenesisBlock() *Block {
	return NewBlock(0, "0", []*txpool.Transaction{}, "genesis")
}

func (bc *Blockchain) AddBlock(transactions []*txpool.Transaction, validator string) {
	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := NewBlock(prevBlock.Index+1, prevBlock.Hash, transactions, validator)
	bc.Blocks = append(bc.Blocks, newBlock)
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
		if block.Header.Number == num {
			return block
		}
	}
	return nil
}

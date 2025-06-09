package blockchain

import txpool "../txpool"

type Blockchain struct {
	Blocks []*Block
}

func NewBlockchain() *Blockchain {
	return &Blockchain{
		Blocks: []*Block{NewGenesisBlock()},
	}
}

func NewGenesisBlock() *Block {
	return NewBlock(0, "0", []txpool.Transaction{}, "genesis")
}

func (bc *Blockchain) AddBlock(transactions []txpool.Transaction, validator string) {
	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := NewBlock(prevBlock.Index+1, prevBlock.Hash, transactions, validator)
	bc.Blocks = append(bc.Blocks, newBlock)
}

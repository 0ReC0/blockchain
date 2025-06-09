package sharding

import (
	"sync"

	blockchain "../blockchain"
	txpool "../txpool"
)

type Shard struct {
	ID     string
	Chain  []*blockchain.Block
	TxPool *txpool.TransactionPool
	mu     sync.Mutex
}

func NewShard(id string) *Shard {
	return &Shard{
		ID:     id,
		Chain:  []*blockchain.Block{blockchain.NewGenesisBlock()},
		TxPool: txpool.NewTransactionPool(),
	}
}

func (s *Shard) AddBlock(block *blockchain.Block) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Chain = append(s.Chain, block)
}

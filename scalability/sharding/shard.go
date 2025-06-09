package sharding

import (
	"time"

	"../../storage/blockchain"
	"../../storage/txpool"
)

type Shard struct {
	ID           string
	Chain        *blockchain.Blockchain
	TxPool       *txpool.TransactionPool
	CurrentBlock *blockchain.Block
}

func NewShard(id string) *Shard {
	return &Shard{
		ID:     id,
		Chain:  blockchain.NewBlockchain(),
		TxPool: txpool.NewTransactionPool(),
		CurrentBlock: &blockchain.Block{
			Index:     0,
			PrevHash:  "0",
			Timestamp: 0,
		},
	}
}

func (s *Shard) AddTransaction(tx *txpool.Transaction) {
	s.TxPool.AddTransaction(tx)
}

func (s *Shard) FinalizeBlock() {
	// Создаем новый блок
	newBlock := &blockchain.Block{
		Index:        s.Chain.Blocks[len(s.Chain.Blocks)-1].Index + 1,
		PrevHash:     s.Chain.Blocks[len(s.Chain.Blocks)-1].Hash,
		Timestamp:    time.Now().Unix(),
		Transactions: s.TxPool.GetTransactions(100),
	}

	newBlock.Hash = newBlock.CalculateHash()
	s.Chain.Blocks = append(s.Chain.Blocks, newBlock)
	s.TxPool.Flush()
}

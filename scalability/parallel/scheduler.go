package parallel

import (
	"blockchain/storage/blockchain"
	"blockchain/storage/txpool"
)

type Scheduler struct {
	Executor *ParallelExecutor
}

func (s *Scheduler) Schedule(pool *txpool.TransactionPool, chain *blockchain.Blockchain) error {
	txs := pool.GetTransactions(100)
	return s.Executor.ExecuteTransactions(txs, chain)
}

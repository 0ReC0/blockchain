package sharding

import (
	"blockchain/storage/blockchain"
	"blockchain/storage/txpool"
)

// Шард
type Shard struct {
	ID        int
	Validators []string
	Chain     *blockchain.Blockchain
	TxPool    *txpool.TransactionPool
}

// Менеджер шардинга
type ShardingManager struct {
	Shards map[int]*Shard
	Router *ShardRouter
}
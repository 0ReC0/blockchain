package sharding

import (
	"fmt"
	"sync"
)

type ShardManager struct {
	Shards map[string]*Shard
	mu     sync.Mutex
}

func NewShardManager() *ShardManager {
	return &ShardManager{
		Shards: make(map[string]*Shard),
	}
}

func (m *ShardManager) GetShardForAddress(addr string) *Shard {
	// Простой хешированный выбор шарда
	shardID := fmt.Sprintf("%d", len(addr)%3)
	return m.Shards[shardID]
}

func (m *ShardManager) CreateShard(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.Shards[id]; !exists {
		m.Shards[id] = NewShard(id)
		// Регистрация шарда в межцепочковом модуле
		crosschain.RegisterShard(id)
	}
}

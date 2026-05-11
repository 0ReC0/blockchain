package sharding

import (
	"blockchain/storage/txpool"
	"sync"
)

// Роутер для определения шарда по транзакции
type ShardRouter struct {
	mu         sync.RWMutex
	ShardCount int
}

// Улучшенная маршрутизация: по хешу адреса получателя для более равномерного распределения
func (r *ShardRouter) RouteTransaction(tx *txpool.Transaction) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.ShardCount <= 0 {
		return 0
	}

	if len(tx.To) == 0 {
		return 0
	}

	// Вычисляем хеш адреса для более равномерного распределения
	hash := 0
	for _, c := range tx.To {
		hash = hash*31 + int(c)
	}

	return hash % r.ShardCount
}

// RouteTransactionBySender маршрутизирует по хешу отправителя
func (r *ShardRouter) RouteTransactionBySender(tx *txpool.Transaction) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.ShardCount <= 0 {
		return 0
	}

	if len(tx.From) == 0 {
		return 0
	}

	// Вычисляем хеш адреса отправителя
	hash := 0
	for _, c := range tx.From {
		hash = hash*31 + int(c)
	}

	return hash % r.ShardCount
}

// RouteByHash маршрутизирует по произвольному хешу (для балансировки)
func (r *ShardRouter) RouteByHash(data string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.ShardCount <= 0 {
		return 0
	}

	hash := 0
	for _, c := range data {
		hash = hash*31 + int(c)
	}

	return hash % r.ShardCount
}

// UpdateShardCount динамически обновляет количество шардов
func (r *ShardRouter) UpdateShardCount(newCount int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if newCount > 0 {
		oldCount := r.ShardCount
		r.ShardCount = newCount
		if newCount != oldCount {
			println("🔄 Router shard count updated:", oldCount, "->", newCount)
		}
	}
}

// GetShardCount возвращает текущее количество шардов
func (r *ShardRouter) GetShardCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ShardCount
}

// ConsistentHashRing - кольцо для консистентного хеширования
// Минимизирует миграцию транзакций при изменении количества шардов
type ConsistentHashRing struct {
	mu          sync.RWMutex
	ring        map[uint64]int // hash -> shard ID
	sortedKeys  []uint64
	replicas    int // количество виртуальных узлов на шард
	shardCount  int
}

// NewConsistentHashRing создаёт новое кольцо для консистентного хеширования
func NewConsistentHashRing(replicas int) *ConsistentHashRing {
	return &ConsistentHashRing{
		ring:       make(map[uint64]int),
		sortedKeys: make([]uint64, 0),
		replicas:   replicas,
	}
}

// AddShard добавляет шард в кольцо
func (c *ConsistentHashRing) AddShard(shardID int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Создаём виртуальные узлы для равномерного распределения
	for i := 0; i < c.replicas; i++ {
		key := c.hashKey(string(rune(shardID)) + string(rune(i)))
		c.ring[key] = shardID
		c.sortedKeys = append(c.sortedKeys, key)
	}

	// Сортируем ключи
	c.sortKeys()
	c.shardCount++
}

// RemoveShard удаляет шард из кольца
func (c *ConsistentHashRing) RemoveShard(shardID int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Удаляем виртуальные узлы
	for i := 0; i < c.replicas; i++ {
		key := c.hashKey(string(rune(shardID)) + string(rune(i)))
		delete(c.ring, key)

		// Удаляем из отсортированного списка
		for j, k := range c.sortedKeys {
			if k == key {
				c.sortedKeys = append(c.sortedKeys[:j], c.sortedKeys[j+1:]...)
				break
			}
		}
	}

	c.shardCount--
}

// GetShard возвращает шард для данного ключа
func (c *ConsistentHashRing) GetShard(key string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.ring) == 0 {
		return 0
	}

	hash := c.hashKey(key)

	// Бинарный поиск ближайшего узла
	idx := c.search(hash)
	return c.ring[c.sortedKeys[idx]]
}

// hashKey вычисляет хеш ключа
func (c *ConsistentHashRing) hashKey(key string) uint64 {
	hash := uint64(0)
	for _, ch := range key {
		hash = hash*31 + uint64(ch)
	}
	return hash
}

// search выполняет бинарный поиск
func (c *ConsistentHashRing) search(hash uint64) int {
	keys := c.sortedKeys
	n := len(keys)

	if n == 0 {
		return 0
	}

	// Бинарный поиск
	left, right := 0, n-1
	for left < right {
		mid := (left + right) / 2
		if keys[mid] < hash {
			left = mid + 1
		} else {
			right = mid
		}
	}

	// Если хеш больше всех ключей, возвращаем первый (кольцо)
	if left == n {
		left = 0
	}

	return left
}

// sortKeys сортирует ключи
func (c *ConsistentHashRing) sortKeys() {
	keys := c.sortedKeys
	n := len(keys)

	// Простая сортировка пузырьком (для небольшого количества ключей)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if keys[j] > keys[j+1] {
				keys[j], keys[j+1] = keys[j+1], keys[j]
			}
		}
	}
}

// GetShardCount возвращает количество шардов в кольце
func (c *ConsistentHashRing) GetShardCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.shardCount
}

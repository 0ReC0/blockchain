package sharding

import (
	"blockchain/storage/txpool"
)

// Роутер для определения шарда по транзакции
type ShardRouter struct {
	ShardCount int
}

// Улучшенная маршрутизация: по хешу адреса получателя для более равномерного распределения
func (r *ShardRouter) RouteTransaction(tx *txpool.Transaction) int {
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

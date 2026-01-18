package sharding

import (
	"blockchain/storage/txpool"
)

// Роутер для определения шарда по транзакции
type ShardRouter struct {
	ShardCount int
}

// Пример маршрутизации: по первому байту адреса получателя
func (r *ShardRouter) RouteTransaction(tx *txpool.Transaction) int {
	if len(tx.To) == 0 {
		return 0
	}
	return int(tx.To[0]) % r.ShardCount
}
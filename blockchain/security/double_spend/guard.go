package double_spend

import (
	"sync"
)

type DoubleSpendGuard struct {
	seenTransactions map[string]bool
	mu               sync.Mutex
}

func NewDoubleSpendGuard() *DoubleSpendGuard {
	return &DoubleSpendGuard{
		seenTransactions: make(map[string]bool),
	}
}

func (g *DoubleSpendGuard) CheckAndMark(txID string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.seenTransactions[txID] {
		return false // двойной расход
	}
	g.seenTransactions[txID] = true
	return true // транзакция уникальна
}

// InitSecurity — инициализирует защиту от двойных трат
func InitSecurity() *DoubleSpendGuard {
	return NewDoubleSpendGuard()
}

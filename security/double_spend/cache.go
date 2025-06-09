package double_spend

import (
	"time"
)

// Очистка кэша каждые N минут
func (g *DoubleSpendGuard) StartCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			<-ticker.C
			g.mu.Lock()
			g.seenTransactions = make(map[string]bool)
			g.mu.Unlock()
		}
	}()
}

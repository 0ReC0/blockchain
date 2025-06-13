package double_spend

import (
	"blockchain/security/audit"
	"sync"
	"time"
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

var auditor *audit.SecurityAuditor

func SetAuditor(a *audit.SecurityAuditor) {
	auditor = a
}

func (g *DoubleSpendGuard) CheckAndMark(txID string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.seenTransactions[txID] {
		auditor.RecordEvent(audit.SecurityEvent{
			Timestamp: time.Now(),
			Type:      "DoubleSpendAttempt",
			Message:   "Detected double spend attempt: " + txID,
			NodeID:    "validator1",
			Severity:  "WARNING",
		})
		return false // двойная трата
	}
	g.seenTransactions[txID] = true
	return true // уникальная транзакция
}

// InitSecurity — инициализирует защиту от двойных трат
func InitSecurity() *DoubleSpendGuard {
	return NewDoubleSpendGuard()
}

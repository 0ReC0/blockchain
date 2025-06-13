package fiftyone

import (
	"blockchain/security/audit"
	"sync"
	"time"
)

type FiftyOnePercentGuard struct {
	ValidatorPower map[string]int64 // адрес -> стейк
	TotalPower     int64
	mu             sync.Mutex
}

func NewFiftyOnePercentGuard(validators map[string]int64) *FiftyOnePercentGuard {
	var total int64
	for _, power := range validators {
		total += power
	}
	return &FiftyOnePercentGuard{
		ValidatorPower: validators,
		TotalPower:     total,
	}
}

var auditor *audit.SecurityAuditor

func SetAuditor(a *audit.SecurityAuditor) {
	auditor = a
}

func (g *FiftyOnePercentGuard) IsMajorityAttackPossible() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, power := range g.ValidatorPower {
		if power*100 > g.TotalPower*51/100 {
			auditor.RecordEvent(audit.SecurityEvent{
				Timestamp: time.Now(),
				Type:      "51PercentAttackRisk",
				Message:   "Validator has >51% stake",
				NodeID:    "validator1",
				Severity:  "CRITICAL",
			})
			return true
		}
	}
	return false
}

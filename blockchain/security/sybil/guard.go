package sybil

import (
	"blockchain/security/audit"
	"crypto/sha256"
	"sync"
	"time"
)

type SybilGuard struct {
	knownNodes     map[string]bool
	validatorNodes map[string]bool
	mu             sync.Mutex
}

func NewSybilGuard(validators []string) *SybilGuard {
	vMap := make(map[string]bool)
	for _, v := range validators {
		vMap[v] = true
	}
	return &SybilGuard{
		knownNodes:     make(map[string]bool),
		validatorNodes: vMap,
	}
}

func (g *SybilGuard) IsKnownNode(id string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.knownNodes[id]
}

var auditor *audit.SecurityAuditor

func SetAuditor(a *audit.SecurityAuditor) {
	auditor = a
}

func (g *SybilGuard) RegisterNode(id string) bool {
	if g.IsValidator(id) {
		return true
	}
	if g.IsKnownNode(id) {
		return true
	}
	hash := sha256.Sum256([]byte(id))
	if hash[0] < 10 {
		g.mu.Lock()
		defer g.mu.Unlock()
		g.knownNodes[id] = true
		return true
	} else {
		auditor.RecordEvent(audit.SecurityEvent{
			Timestamp: time.Now(),
			Type:      "SybilNodeRejected",
			Message:   "Sybil node registration rejected: " + id,
			NodeID:    "validator1",
			Severity:  "WARNING",
		})
		return false
	}
}

func (g *SybilGuard) IsValidator(id string) bool {
	return g.validatorNodes[id]
}

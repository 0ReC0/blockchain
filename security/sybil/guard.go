package sybil

import (
	"crypto/sha256"
	"sync"
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

func (g *SybilGuard) RegisterNode(id string) bool {
	if g.IsValidator(id) {
		return true // валидаторы всегда доверяются
	}
	if g.IsKnownNode(id) {
		return true
	}
	// Простая проверка (можно заменить на PoW, PoS, Web of Trust)
	hash := sha256.Sum256([]byte(id))
	if hash[0] < 10 { // искусственное ограничение
		g.mu.Lock()
		defer g.mu.Unlock()
		g.knownNodes[id] = true
		return true
	}
	return false
}

func (g *SybilGuard) IsValidator(id string) bool {
	return g.validatorNodes[id]
}

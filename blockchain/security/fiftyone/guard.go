package fiftyone

import (
	"sync"
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

func (g *FiftyOnePercentGuard) IsMajorityAttackPossible() bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Проверяем, есть ли у одного узла более 51%
	for _, power := range g.ValidatorPower {
		if power*100 > g.TotalPower*51/100 {
			return true
		}
	}
	return false
}

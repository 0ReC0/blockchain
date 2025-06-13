package pos

// выбор валидатора

import (
	"math/rand"
	"time"

	"blockchain/governance/reputation"
)

type ValidatorPool []*Validator

func (p ValidatorPool) Select() *Validator {
	if len(p) == 0 {
		return nil
	}

	repModule := reputation.NewReputationSystem()

	totalWeight := 0.0
	weights := make([]float64, len(p))
	addresses := make([]string, len(p))
	balances := make([]int64, len(p))

	for i, v := range p {
		repScore := repModule.CalculateScore(v.Address, true)
		weight := float64(v.Balance) * repScore
		totalWeight += weight
		weights[i] = totalWeight
		addresses[i] = v.Address
		balances[i] = v.Balance
	}

	if totalWeight == 0 {
		return &Validator{
			ID:      p[0].ID,
			Address: p[0].Address,
			Balance: p[0].Balance,
		}
	}

	rand.Seed(time.Now().UnixNano())
	r := rand.Float64() * totalWeight

	for i, w := range weights {
		if r <= w {
			return &Validator{
				ID:      p[i].ID,
				Address: addresses[i],
				Balance: balances[i],
			}
		}
	}

	return p[0]
}

func NewValidatorPool(validators []*Validator) *ValidatorPool {
	vp := ValidatorPool(validators)
	return &vp
}

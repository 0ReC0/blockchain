package reputation

import (
	"math/rand"
	"time"
)

// Validator представляет валидатора
type Validator struct {
	Address string
	Stake   float64
}

// SelectValidator выбирает валидатора на основе: Stake * Reputation.Weight
func SelectValidator(validators []Validator, repSystem *ReputationSystem) Validator {
	rand.Seed(time.Now().UnixNano())

	totalWeight := 0.0
	weights := make([]float64, len(validators))
	addresses := make([]string, len(validators))
	stakes := make([]float64, len(validators))

	for i, v := range validators {
		repScore := repSystem.CalculateScore(v.Address, true)
		weight := v.Stake * repScore
		totalWeight += weight
		weights[i] = totalWeight
		addresses[i] = v.Address
		stakes[i] = v.Stake
	}

	if totalWeight == 0 {
		// Резервный случай
		i := rand.Intn(len(validators))
		return validators[i]
	}

	r := rand.Float64() * totalWeight

	for i, w := range weights {
		if r <= w {
			return Validator{
				Address: addresses[i],
				Stake:   stakes[i],
			}
		}
	}

	return validators[rand.Intn(len(validators))]
}

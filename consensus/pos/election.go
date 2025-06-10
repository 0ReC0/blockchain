package pos

// выбор валидатора

import (
	"math/rand"
	"time"

	"../../governance/reputation"
)

type ValidatorPool []*Validator

func (p ValidatorPool) Select() *Validator {
	if len(p) == 0 {
		return nil
	}

	// Использование репутации для выбора валидатора
	repModule := reputation.NewReputationModule()

	// Обновляем репутацию перед выбором
	for _, v := range p {
		repModule.UpdateReputation(v.Address, 1.0)
	}

	rand.Seed(time.Now().UnixNano())
	total := 0.0
	for _, v := range p {
		total += v.Weight() * repModule.CalculateScore(v.Address, true)
	}

	r := rand.Float64() * total
	for _, v := range p {
		r -= v.Weight()
		if r <= 0 {
			return v
		}
	}
	return p[0]
}

func NewValidatorPool(validators []*Validator) *ValidatorPool {
	vp := ValidatorPool(validators)
	return &vp
}

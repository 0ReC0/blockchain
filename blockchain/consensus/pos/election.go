package pos

// выбор валидатора

import (
	"math/rand"
	"time"

	"blockchain/governance/reputation"
)

type ValidatorPool []*Validator

func (p ValidatorPool) Select(round int64) *Validator {
	if len(p) == 0 {
		return nil
	}

	// Инициализируем систему репутации
	repModule := reputation.NewReputationSystem()

	// Используем раунд в сид случайности для предсказуемости и справедливости
	rand.Seed(time.Now().UnixNano() + int64(round))

	// Подсчёт общего веса
	totalWeight := 0.0
	weights := make([]float64, len(p))
	validators := make([]*Validator, len(p))

	for i, v := range p {
		repScore := repModule.CalculateScore(v.Address, true)

		// Вес = баланс * репутация
		weight := float64(v.Balance) * repScore
		totalWeight += weight

		weights[i] = totalWeight
		validators[i] = v
	}

	// Если весь вес нулевой — возвращаем первого
	if totalWeight == 0 {
		return p[0]
	}

	// Генерируем случайное число в диапазоне [0, totalWeight)
	r := rand.Float64() * totalWeight

	// Находим валидатора
	for i, w := range weights {
		if r <= w {
			return validators[i]
		}
	}

	return validators[0]
}

func NewValidatorPool(validators []*Validator) *ValidatorPool {
	vp := ValidatorPool(validators)
	return &vp
}
func (p ValidatorPool) GetValidatorByAddress(addr string) *Validator {
	for _, v := range p {
		if v.Address == addr {
			return v
		}
	}
	return nil
}

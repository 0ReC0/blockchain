package pos

// выбор валидатора

import (
	"math/rand"
	"time"
)

type ValidatorPool []*Validator

func (p ValidatorPool) Select() *Validator {
	if len(p) == 0 {
		return nil
	}

	// Простой рандомизированный выбор по ставке
	rand.Seed(time.Now().UnixNano())
	total := 0.0
	for _, v := range p {
		total += v.Weight()
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

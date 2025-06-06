package pos

// валидатор

type Validator struct {
	Address string
	Stake   int64
}

func NewValidator(addr string, stake int64) *Validator {
	return &Validator{Address: addr, Stake: stake}
}

func (v *Validator) Weight() float64 {
	return float64(v.Stake)
}

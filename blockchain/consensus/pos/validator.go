package pos

// валидатор

type Validator struct {
	ID      string
	Address string
	Balance int64
}

func NewValidatorWithAddress(id, address string, balance int64) *Validator {
	return &Validator{
		ID:      id,
		Address: address,
		Balance: balance,
	}
}

func (v *Validator) Weight() float64 {
	return float64(v.Balance) // ✅ Используем Balance
}

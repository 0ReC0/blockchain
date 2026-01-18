package pos

// валидатор

type Validator struct {
	ID             string
	Address        string
	Balance        int64
	CommissionEarned int64  // Добавлено: сумма заработанных комиссий
}

func NewValidatorWithAddress(id, address string, balance int64) *Validator {
	return &Validator{
		ID:      id,
		Address: address,
		Balance: balance,
	}
}

func (v *Validator) Weight() float64 {
	// Обновлено: добавлен учет комиссий в вес валидатора
	return float64(v.Balance + v.CommissionEarned)
}

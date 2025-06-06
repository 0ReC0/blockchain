package pos

// модель ставок

type Stake struct {
	Address string
	Amount  int64
}

func NewStake(addr string, amount int64) *Stake {
	return &Stake{Address: addr, Amount: amount}
}

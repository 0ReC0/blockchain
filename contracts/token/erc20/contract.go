package erc20

type Token struct {
	Name_        string
	Symbol_      string
	Decimals_    int
	TotalSupply_ uint64
	Balances     map[string]uint64
	Allowances   map[string]map[string]uint64
}

func NewToken(name, symbol string, decimals int, totalSupply uint64) *Token {
	return &Token{
		Name_:        name,
		Symbol_:      symbol,
		Decimals_:    decimals,
		TotalSupply_: totalSupply,
		Balances:     make(map[string]uint64),
		Allowances:   make(map[string]map[string]uint64),
	}
}

func (t *Token) Name() string {
	return t.Name_
}

func (t *Token) Symbol() string {
	return t.Symbol_
}

func (t *Token) Decimals() int {
	return t.Decimals_
}

func (t *Token) TotalSupply() uint64 {
	return t.TotalSupply_
}

func (t *Token) BalanceOf(address string) uint64 {
	return t.Balances[address]
}

func (t *Token) Transfer(from, to string, amount uint64) bool {
	if t.Balances[from] < amount {
		return false
	}
	t.Balances[from] -= amount
	t.Balances[to] += amount
	return true
}

func (t *Token) Approve(owner, spender string, amount uint64) bool {
	if t.Allowances[owner] == nil {
		t.Allowances[owner] = make(map[string]uint64)
	}
	t.Allowances[owner][spender] = amount
	return true
}

func (t *Token) Allowance(owner, spender string) uint64 {
	return t.Allowances[owner][spender]
}

func (t *Token) TransferFrom(from, to string, amount uint64) bool {
	if t.Allowances[from][spender] < amount {
		return false
	}
	t.Balances[from] -= amount
	t.Balances[to] += amount
	t.Allowances[from][spender] -= amount
	return true
}

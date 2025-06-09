package erc1155

type MultiToken struct {
	Balances  map[string]map[uint64]uint64
	Approvals map[string]bool
}

func NewMultiToken() *MultiToken {
	return &MultiToken{
		Balances:  make(map[string]map[uint64]uint64),
		Approvals: make(map[string]bool),
	}
}

func (m *MultiToken) BalanceOf(owner string, tokenId uint64) uint64 {
	return m.Balances[owner][tokenId]
}

func (m *MultiToken) BalanceOfBatch(owners []string, tokenIds []uint64) []uint64 {
	var result []uint64
	for i := 0; i < len(owners); i++ {
		result = append(result, m.BalanceOf(owners[i], tokenIds[i]))
	}
	return result
}

func (m *MultiToken) Transfer(from, to string, tokenId uint64, amount uint64) {
	if m.Balances[from][tokenId] < amount {
		return
	}
	m.Balances[from][tokenId] -= amount
	m.Balances[to][tokenId] += amount
}

func (m *MultiToken) TransferBatch(from string, to string, tokenIds []uint64, amounts []uint64) {
	for i := 0; i < len(tokenIds); i++ {
		m.Transfer(from, to, tokenIds[i], amounts[i])
	}
}

func (m *MultiToken) SetApprovalForAll(operator string, approved bool) {
	m.Approvals[operator] = approved
}

func (m *MultiToken) IsApprovedForAll(owner, operator string) bool {
	return m.Approvals[operator]
}

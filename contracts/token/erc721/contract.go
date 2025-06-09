package erc721

type NFT struct {
	Name_     string
	Symbol_   string
	Balances  map[string]uint64
	Owners    map[uint64]string
	Approvals map[uint64]string
}

func NewNFT(name, symbol string) *NFT {
	return &NFT{
		Name_:     name,
		Symbol_:   symbol,
		Balances:  make(map[string]uint64),
		Owners:    make(map[uint64]string),
		Approvals: make(map[uint64]string),
	}
}

func (n *NFT) Name() string {
	return n.Name_
}

func (n *NFT) Symbol() string {
	return n.Symbol_
}

func (n *NFT) BalanceOf(owner string) uint64 {
	return n.Balances[owner]
}

func (n *NFT) OwnerOf(tokenId uint64) string {
	return n.Owners[tokenId]
}

func (n *NFT) TransferFrom(from, to string, tokenId uint64) {
	if n.Owners[tokenId] != from {
		return
	}
	n.Balances[from]--
	n.Balances[to]++
	n.Owners[tokenId] = to
}

func (n *NFT) Approve(to string, tokenId uint64) {
	n.Approvals[tokenId] = to
}

func (n *NFT) GetApproved(tokenId uint64) string {
	return n.Approvals[tokenId]
}

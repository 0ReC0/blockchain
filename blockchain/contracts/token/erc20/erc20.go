package erc20

type ERC20 interface {
	Name() string
	Symbol() string
	Decimals() int
	TotalSupply() uint64
	BalanceOf(address string) uint64
	Transfer(from, to string, amount uint64) bool
	Approve(owner, spender string, amount uint64) bool
	Allowance(owner, spender string) uint64
	TransferFrom(from, to string, amount uint64) bool
}

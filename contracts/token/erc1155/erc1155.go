package erc1155

type ERC1155 interface {
	BalanceOf(owner string, tokenId uint64) uint64
	BalanceOfBatch(owners []string, tokenIds []uint64) []uint64
	Transfer(from, to string, tokenId uint64, amount uint64)
	TransferBatch(from string, to string, tokenIds []uint64, amounts []uint64)
	SetApprovalForAll(operator string, approved bool)
	IsApprovedForAll(owner, operator string) bool
}

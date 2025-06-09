package erc721

type ERC721 interface {
	Name() string
	Symbol() string
	BalanceOf(owner string) uint64
	OwnerOf(tokenId uint64) string
	TransferFrom(from, to string, tokenId uint64)
	Approve(to string, tokenId uint64)
	GetApproved(tokenId uint64) string
}

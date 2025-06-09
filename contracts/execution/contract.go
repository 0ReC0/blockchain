package execution

import (
	"fmt"

	"../token/erc1155"
	"../token/erc20"
	"../token/erc721"
)

type Contract interface {
	Execute(input []byte) ([]byte, error)
}

type ContractHandler struct {
	ERC20Contracts   map[string]*erc20.Token
	ERC721Contracts  map[string]*erc721.NFT
	ERC1155Contracts map[string]*erc1155.MultiToken
}

func NewContractHandler() *ContractHandler {
	return &ContractHandler{
		ERC20Contracts:   make(map[string]*erc20.Token),
		ERC721Contracts:  make(map[string]*erc721.NFT),
		ERC1155Contracts: make(map[string]*erc1155.MultiToken),
	}
}

func (h *ContractHandler) DeployERC20(name, symbol string, decimals int, totalSupply uint64) string {
	token := erc20.NewToken(name, symbol, decimals, totalSupply)
	addr := generateAddress()
	h.ERC20Contracts[addr] = token
	return addr
}

func (h *ContractHandler) CallERC20(addr string, method string, args ...interface{}) (interface{}, error) {
	token, exists := h.ERC20Contracts[addr]
	if !exists {
		return nil, fmt.Errorf("contract not found")
	}

	switch method {
	case "transfer":
		if len(args) != 3 {
			return nil, fmt.Errorf("invalid args")
		}
		from := args[0].(string)
		to := args[1].(string)
		amount := args[2].(uint64)
		return token.Transfer(from, to, amount), nil
		// другие методы...
	}
	return nil, fmt.Errorf("method not found")
}

package shielded

import (
	"../../storage/blockchain"
	"../private_tx"
)

type ShieldedBlock struct {
	*blockchain.Block
	PrivateTransactions []*private_tx.PrivateTransaction
}

func NewShieldedBlock(baseBlock *blockchain.Block, privateTxs []*private_tx.PrivateTransaction) *ShieldedBlock {
	return &ShieldedBlock{
		Block:               baseBlock,
		PrivateTransactions: privateTxs,
	}
}

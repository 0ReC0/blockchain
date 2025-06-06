package bft

// реализация BFT (упрощённый Tendermint)

import (
	"fmt"
	"time"
)

type BFTNode struct {
	Address   string
	Validator *pos.Validator
	Height    int64
	Round     int64
}

func NewBFTNode(addr string, val *pos.Validator) *BFTNode {
	return &BFTNode{
		Address:   addr,
		Validator: val,
		Height:    0,
		Round:     0,
	}
}

func (n *BFTNode) Start() {
	for {
		n.RunConsensusRound()
		n.Height++
	}
}

func (n *BFTNode) RunConsensusRound() {
	validators := pos.ValidatorPool{
		n.Validator,
		{Address: "validator2", Stake: 1000},
		{Address: "validator3", Stake: 1500},
	}

	proposer := validators.Select()
	round := NewRound(n.Height, n.Round, proposer.Address)

	fmt.Printf("Starting round %d for height %d. Proposer: %s\n", n.Round, n.Height, proposer.Address)

	// 1. Propose
	if proposer.Address == n.Address {
		block := []byte("block-data")
		round.ProposedBlock = block
		round.Step = MsgPropose
		fmt.Printf("Proposing block: %x\n", block)
	}

	time.Sleep(1 * time.Second)

	// 2. Prevote
	round.Prevotes[n.Address] = []byte("prevote")
	round.Step = MsgPrevote
	fmt.Printf("Prevote from %s\n", n.Address)

	time.Sleep(1 * time.Second)

	// 3. Precommit
	round.Precommits[n.Address] = []byte("precommit")
	round.Step = MsgPrecommit
	fmt.Printf("Precommit from %s\n", n.Address)

	time.Sleep(1 * time.Second)

	// 4. Commit
	if len(round.Precommits) >= 2 {
		round.Step = MsgCommit
		fmt.Printf("Block committed: %x\n", round.ProposedBlock)
	}
}

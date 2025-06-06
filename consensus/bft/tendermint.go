package bft

// реализация BFT (упрощённый Tendermint)

import (
	"fmt"
	"time"

	"../../network/gossip"
	"../pos"
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

func (n *BFTNode) BroadcastMessage(msgType MessageType, data []byte) {
	msg := &gossip.ConsensusMessage{
		Type:   msgType,
		Height: n.Height,
		Round:  n.Round,
		From:   n.Address,
		Data:   data,
	}

	gossip.BroadcastConsensusMessage(n.ValidatorPool, msg)
}

func (n *BFTNode) RunConsensusRound() {
	// Выбор пропосера
	proposer := n.ValidatorPool.Select()
	round := NewRound(n.Height, n.Round, proposer.Address)

	fmt.Printf("Starting round %d for height %d. Proposer: %s\n", n.Round, n.Height, proposer.Address)

	// 1. Propose
	if proposer.Address == n.Address {
		block := []byte("block-data")
		round.ProposedBlock = block
		n.BroadcastMessage(MsgPropose, block)
	}

	time.Sleep(1 * time.Second)

	// 2. Prevote
	round.Prevotes[n.Address] = []byte("prevote")
	n.BroadcastMessage(MsgPrevote, []byte("prevote"))
	fmt.Printf("Prevote from %s\n", n.Address)

	time.Sleep(1 * time.Second)

	// 3. Precommit
	round.Precommits[n.Address] = []byte("precommit")
	n.BroadcastMessage(MsgPrecommit, []byte("precommit"))
	fmt.Printf("Precommit from %s\n", n.Address)

	time.Sleep(1 * time.Second)

	// 4. Commit
	if len(round.Precommits) >= 2 {
		n.BroadcastMessage(MsgCommit, round.ProposedBlock)
		fmt.Printf("Block committed: %x\n", round.ProposedBlock)
	}
}

package bft

import "blockchain/network/gossip"

// раунд консенсуса

// Round — структура раунда консенсуса
type Round struct {
	Height        int64
	Round         int64
	Step          gossip.MessageType
	Proposer      string
	ProposedBlock []byte
	Prevotes      map[string][]byte
	Precommits    map[string][]byte
}

func NewRound(height, round int64, proposer string) *Round {
	return &Round{
		Height:     height,
		Round:      round,
		Step:       gossip.StatePropose,
		Proposer:   proposer,
		Prevotes:   make(map[string][]byte),
		Precommits: make(map[string][]byte),
	}
}

package manager

import (
	"fmt"
)

type ConsensusType string

const (
	ConsensusPoS ConsensusType = "PoS"
	ConsensusBFT ConsensusType = "BFT"
)

type ConsensusSwitcher struct {
	Type ConsensusType
}

func NewConsensusSwitcher(t ConsensusType) *ConsensusSwitcher {
	return &ConsensusSwitcher{Type: t}
}

func (cs *ConsensusSwitcher) StartConsensus() {
	switch cs.Type {
	case ConsensusPoS:
		fmt.Println("Starting PoS consensus...")
		// Запуск PoS
	case ConsensusBFT:
		fmt.Println("Starting BFT consensus...")
		// Запуск BFT
	default:
		fmt.Println("Unknown consensus type")
	}
}

package manager

import (
	"fmt"
	"time"

	"../../security/fiftyone"
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
		validatorsMap := map[string]int64{
			"validator1": 2000,
			"validator2": 1000,
		}
		fiftyOneGuard := fiftyone.NewFiftyOnePercentGuard(validatorsMap)
		fiftyOneGuard.Monitor(30 * time.Second)
		// Запуск PoS
	case ConsensusBFT:
		fmt.Println("Starting BFT consensus...")
		// Запуск BFT
	default:
		fmt.Println("Unknown consensus type")
	}
}

func (cs *ConsensusSwitcher) Run() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		cs.StartConsensus()
	}
}

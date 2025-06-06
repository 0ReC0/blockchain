package manager

import (
	"time"
)

type ConsensusManager struct {
	Switcher *ConsensusSwitcher
}

func NewConsensusManager(t ConsensusType) *ConsensusManager {
	return &ConsensusManager{
		Switcher: NewConsensusSwitcher(t),
	}
}

func (cm *ConsensusManager) Run() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		cm.Switcher.StartConsensus()
	}
}

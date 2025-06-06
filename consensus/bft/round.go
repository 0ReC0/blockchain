package bft

// раунд консенсуса

type Round struct {
	Height        int64
	Round         int64
	Step          MessageType
	Proposer      string
	ProposedBlock []byte
	Prevotes      map[string][]byte
	Precommits    map[string][]byte
}

func NewRound(height, round int64, proposer string) *Round {
	return &Round{
		Height:     height,
		Round:      round,
		Step:       MsgPropose,
		Proposer:   proposer,
		Prevotes:   make(map[string][]byte),
		Precommits: make(map[string][]byte),
	}
}

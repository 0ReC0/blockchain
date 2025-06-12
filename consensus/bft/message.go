package bft

// типы сообщений BFT

type ConsensusState string

const (
	StatePropose   ConsensusState = "propose"
	StatePrevote   ConsensusState = "prevote"
	StatePrecommit ConsensusState = "precommit"
	StateCommit    ConsensusState = "commit"
)

type Message struct {
	Type      ConsensusState
	Height    int64
	Round     int64
	Proposer  string
	Data      []byte
	Signature []byte
}

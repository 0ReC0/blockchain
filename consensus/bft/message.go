package bft

// типы сообщений BFT

type MessageType string

const (
	MsgPropose   MessageType = "propose"
	MsgPrevote   MessageType = "prevote"
	MsgPrecommit MessageType = "precommit"
	MsgCommit    MessageType = "commit"
)

type Message struct {
	Type      MessageType
	Height    int64
	Round     int64
	Proposer  string
	Data      []byte
	Signature []byte
}

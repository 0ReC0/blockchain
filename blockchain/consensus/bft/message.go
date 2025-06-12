package bft

import "blockchain/network/gossip"

// типы сообщений BFT

type Message struct {
	Type      gossip.MessageType
	Height    int64
	Round     int64
	Proposer  string
	Data      []byte
	Signature []byte
}

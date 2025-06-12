package p2p

import (
	"encoding/json"
	"time"
)

type Handshake struct {
	NodeID    string
	Timestamp int64
	UserAgent string
	Protocols []string
}

func NewHandshake(nodeID string) *Handshake {
	return &Handshake{
		NodeID:    nodeID,
		Timestamp: time.Now().Unix(),
		UserAgent: "blockchain-node/1.0",
		Protocols: []string{"gossip/1.0", "rpc/1.0"},
	}
}

func (h *Handshake) Serialize() ([]byte, error) {
	return json.Marshal(h)
}

func DeserializeHandshake(data []byte) (*Handshake, error) {
	var h Handshake
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, err
	}
	return &h, nil
}

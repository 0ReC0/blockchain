package p2p

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
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

func (n *Node) PerformHandshake(conn *tls.Conn) error {
	hs := NewHandshake(n.ID)
	data, _ := hs.Serialize()
	_, err := conn.Write(data)
	if err != nil {
		return err
	}

	// Read response
	buf := make([]byte, 1024)
	bytesRead, err := conn.Read(buf)
	remoteHS, err := DeserializeHandshake(buf[:bytesRead])
	if err != nil {
		return err
	}
	fmt.Printf("Handshake with %s successful\n", remoteHS.NodeID)
	return nil
}

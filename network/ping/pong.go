package ping

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"
)

type Ping struct {
	Timestamp int64  `json:"timestamp"`
	NodeID    string `json:"node_id"`
}

func (p *Ping) Serialize() ([]byte, error) {
	return json.Marshal(p)
}

func DeserializePing(data []byte) (*Ping, error) {
	var p Ping
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func SendPing(conn *tls.Conn, nodeID string) error {
	ping := &Ping{
		Timestamp: time.Now().UnixNano(),
		NodeID:    nodeID,
	}
	data, _ := ping.Serialize()
	_, err := conn.Write(data)
	if err != nil {
		return err
	}
	fmt.Printf("Sent ping to %s\n", nodeID)
	return nil
}

func HandlePing(conn *tls.Conn, data []byte) {
	ping, err := DeserializePing(data)
	if err != nil {
		return
	}
	fmt.Printf("Received ping from %s at %d\n", ping.NodeID, ping.Timestamp)

	// Отправляем pong
	pong := &Ping{
		Timestamp: time.Now().UnixNano(),
		NodeID:    "node1",
	}
	pongData, _ := pong.Serialize()
	_, err = conn.Write(pongData)
	if err != nil {
		fmt.Printf("Failed to send pong: %v\n", err)
	}
}

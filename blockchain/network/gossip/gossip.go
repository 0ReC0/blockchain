package gossip

// протокол рассылки

import (
	"bytes"
	"crypto/tls"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"

	"blockchain/network/p2p"
	"blockchain/network/peer"
)

func (m *GossipMessage) Encode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(m)
	return buf.Bytes(), err
}

func DecodeMessage(data []byte) (*GossipMessage, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	var msg GossipMessage
	err := dec.Decode(&msg)
	return &msg, err
}

func Broadcast(peers []*peer.Peer, msg *GossipMessage) error {
	for _, peer := range peers {
		conn, err := net.Dial("tcp", peer.Addr)
		if err != nil {
			fmt.Printf("Can't connect to peer %s: %v\n", peer.ID, err)
			continue
		}
		defer conn.Close()
		encoded, _ := msg.Encode()
		_, err = conn.Write(encoded)
		if err != nil {
			return err
		}
	}
	return nil
}

// BroadcastSignedConsensusMessage — рассылает подписанные сообщения всем пирам
func BroadcastSignedConsensusMessage(peers []*peer.Peer, msg *SignedConsensusMessage) error {
	for _, peer := range peers {
		conn, err := tls.Dial("tcp", peer.Addr, p2p.GenerateClientTLSConfig())
		if err != nil {
			fmt.Printf("Can't connect to peer %s: %v\n", peer.ID, err)
			continue
		}
		defer conn.Close()

		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(msg); err != nil {
			return err
		}
	}
	return nil
}

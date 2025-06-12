package gossip

import (
	"crypto/tls"
	"encoding/json"
	"fmt"

	"blockchain/crypto/signature"
	"blockchain/network/p2p"
	"blockchain/network/peer"
)

type ConsensusMessage struct {
	Type   MessageType
	Height int64
	Round  int64
	From   string
	Data   []byte
}

func (m *ConsensusMessage) Encode() ([]byte, error) {
	return json.Marshal(m)
}

func DecodeConsensusMessage(data []byte) (*ConsensusMessage, error) {
	var msg ConsensusMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func HandleSignedMessage(data []byte) (*ConsensusMessage, error) {
	msg, err := DecodeSignedMessage(data)
	if err != nil {
		return nil, err
	}

	// 1. Проверяем подпись
	pubKey, err := signature.LoadPublicKey(msg.From)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key for %s: %v", msg.From, err)
	}

	if !signature.Verify(pubKey, msg.Data, msg.Signature) {
		return nil, fmt.Errorf("invalid signature from %s", msg.From)
	}

	return &ConsensusMessage{
		Type:   msg.Type,
		Height: msg.Height,
		Round:  msg.Round,
		From:   msg.From,
		Data:   msg.Data,
	}, nil
}

func BroadcastConsensusMessage(peers []*peer.Peer, msg *ConsensusMessage) error {
	for _, peer := range peers {
		conn, err := tls.Dial("tcp", peer.Addr, p2p.GenerateTLSConfig())
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

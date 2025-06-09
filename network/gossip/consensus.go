package gossip

import (
	"crypto/tls"
	"encoding/json"
	"fmt"

	"../crypto"
	"../peer"
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

	// Проверяем подпись
	if !signature.VerifySignature(msg.From, msg.Data, msg.Signature) {
		return nil, fmt.Errorf("invalid signature from %s", msg.From)
	}

	return msg, nil
}

func BroadcastConsensusMessage(peers []*peer.Peer, msg *ConsensusMessage) error {
	for _, peer := range peers {
		conn, err := tls.Dial("tcp", peer.Addr, crypto.GenerateTLSConfig())
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

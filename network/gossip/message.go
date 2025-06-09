package gossip

// типы сообщений
import (
	"encoding/json"
)

type MessageType string

const (
	MsgBlock   MessageType = "block"
	MsgTx      MessageType = "tx"
	MsgStatus  MessageType = "status"
	MsgRequest MessageType = "request"
)

type Message struct {
	Type MessageType
	From string
	Data []byte
}

// SignedConsensusMessage — структура сообщения с подписью
type SignedConsensusMessage struct {
	Type      MessageType `json:"type"`
	Height    int64       `json:"height"`
	Round     int64       `json:"round"`
	From      string      `json:"from"`
	Data      []byte      `json:"data"`
	Signature []byte      `json:"signature"`
}

// DecodeSignedMessage — декодирует байты в SignedConsensusMessage
func DecodeSignedMessage(data []byte) (*SignedConsensusMessage, error) {
	var msg SignedConsensusMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

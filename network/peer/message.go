package gossip

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

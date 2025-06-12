package p2p

// точка входа узла

import (
	"crypto/tls"
	"fmt"
	"log"

	"blockchain/consensus/bft"
	"blockchain/network/gossip"
	"blockchain/network/peer"
)

type Node struct {
	ID      string
	Addr    string
	PeerMgr *peer.PeerManager
}

func NewNode(id, addr string) *Node {
	return &Node{
		ID:      id,
		Addr:    addr,
		PeerMgr: peer.NewPeerManager(),
	}
}

func (n *Node) Start() {
	fmt.Printf("Node %s started at %s\n", n.ID, n.Addr)
	go n.listenTLS()
}
func (n *Node) listenTLS() {
	config := GenerateTLSConfig()
	listener, err := tls.Listen("tcp", n.Addr, config)
	if err != nil {
		log.Fatalf("Failed to start TLS listener: %v", err)
	}
	for {
		rawConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		// Приводим net.Conn к *tls.Conn
		tlsConn, ok := rawConn.(*tls.Conn)
		if !ok {
			log.Println("Connection is not a TLS connection")
			rawConn.Close()
			continue
		}

		go n.handleSecureConnection(tlsConn)
	}
}

func (n *Node) handleSecureConnection(conn *tls.Conn) {
	defer conn.Close()

	if err := n.PerformHandshake(conn); err != nil {
		fmt.Printf("Handshake failed: %v\n", err)
		return
	}

	for {
		buf := make([]byte, 4096)
		nBytes, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("Connection closed: %v\n", err)
			return
		}

		msg, err := gossip.DecodeConsensusMessage(buf[:nBytes])
		if err != nil {
			// Обработка сообщений консенсуса
			go n.handleConsensusMessage(msg)
			return
		}

		switch msg.Type {
		case gossip.MsgPing:
			n.handlePing(conn, msg)
		case gossip.MsgPong:
			fmt.Printf("Received pong from %s\n", msg.From)
		default:
			fmt.Printf("Received message from %s: %s\n", msg.From, msg.Type)
		}
	}
}
func (n *Node) handleConsensusMessage(msg *gossip.ConsensusMessage) {
	// Передача сообщений консенсуса в модуль BFT
	bftHandler := bft.NewBFTMessageHandler(n.PeerMgr)
	bftHandler.ProcessMessage(msg)
}

func (n *Node) handlePing(conn *tls.Conn, msg *gossip.ConsensusMessage) {
	fmt.Printf("Received ping from %s\n", msg.From)

	// Отправляем pong
	pong := &gossip.ConsensusMessage{
		Type:   gossip.MsgPong,
		From:   n.ID,
		Height: msg.Height,
		Round:  msg.Round,
	}
	data, _ := pong.Encode()

	_, err := conn.Write(data)
	if err != nil {
		fmt.Printf("Failed to send pong: %v\n", err)
	}
}

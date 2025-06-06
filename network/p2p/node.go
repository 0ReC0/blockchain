package p2p

// точка входа узла

import (
	"crypto/tls"
	"fmt"

	"../gossip"
	"../peer"
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
	listener, _ := tls.Listen("tcp", n.Addr, config)
	for {
		conn, _ := listener.Accept()
		go n.handleSecureConnection(conn)
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
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("Connection closed: %v\n", err)
			return
		}

		msg, err := gossip.DecodeMessage(buf[:n])
		if err != nil {
			continue
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

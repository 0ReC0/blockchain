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
	if err := n.PerformHandshake(conn); err != nil {
		fmt.Printf("Handshake failed: %v\n", err)
		return
	}

	buf := make([]byte, 4096)
	n, _ := conn.Read(buf)
	msg, _ := gossip.DecodeMessage(buf[:n])
	fmt.Printf("Secure message from %s: %s\n", msg.From, msg.Type)
}

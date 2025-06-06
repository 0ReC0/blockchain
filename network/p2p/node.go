package p2p

// точка входа узла

import (
	"fmt"
	"net"

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
	go n.listen()
}

func (n *Node) listen() {
	listener, _ := net.Listen("tcp", n.Addr)
	for {
		conn, _ := listener.Accept()
		go n.handleConnection(conn)
	}
}

func (n *Node) handleConnection(conn net.Conn) {
	buf := make([]byte, 4096)
	n, _ := conn.Read(buf)
	msg, _ := gossip.DecodeMessage(buf[:n])
	fmt.Printf("Received message from %s: %s\n", msg.From, msg.Type)
}

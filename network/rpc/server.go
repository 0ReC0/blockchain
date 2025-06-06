package p2p

import (
	"blockchain/network/gossip"
	"blockchain/network/peer"
)

func StartNetwork() {
	node := NewNode("node1", ":3000")
	node.PeerMgr.AddPeer(peer.NewPeer("peer1", ":3001"))
	node.Start()

	// Отправка тестового блока
	msg := &gossip.Message{
		Type: gossip.MsgBlock,
		From: node.ID,
		Data: []byte("block-123"),
	}
	gossip.Broadcast(node.PeerMgr.GetPeers(), msg)
}

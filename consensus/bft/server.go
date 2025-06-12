package bft

// запуск узла

import (
	"blockchain/network/gossip"
	"blockchain/network/peer"
)

func StartNetwork() {
	node := NewNode("node1", ":3000")
	node.PeerMgr.AddPeer(peer.NewPeer("peer1", ":3001"))
	node.Start()

	go peer.BroadcastPresence(node.Addr)
	go peer.ListenForPeers()

	// Отправка тестового блока
	msg := &gossip.GossipMessage{
		Type: gossip.MsgBlock,
		From: node.ID,
		Data: []byte("block-123"),
	}
	gossip.Broadcast(node.PeerMgr.GetPeers(), msg)
}

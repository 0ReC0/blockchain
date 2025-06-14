package bft

import (
	"blockchain/network/gossip"
	"blockchain/network/peer"
	"blockchain/storage/blockchain"
	"blockchain/storage/txpool"
)

// StartNetwork запускает узел BFT с переданными txPool и chain
func StartNetwork(
	txPool *txpool.TransactionPool,
	chain *blockchain.Blockchain,
	nodeID string,
	nodePort string,
) {
	node := NewNode(nodeID, nodePort, txPool, chain)

	// Добавляем пир (для примера)
	node.PeerMgr.AddPeer(peer.NewPeer("peer1", ":3001", nil))

	// Запуск узла
	node.Start()

	// === Добавьте инициализацию UDP перед запуском горутин ===
	peer.InitUDPSocket(nodePort) // инициализируем UDP-сокет на порту узла

	// Горутины для сети
	go peer.BroadcastPresence(node.Addr)
	go peer.ListenForPeers() // теперь это безопасно

	// Отправка тестового блока
	msg := &gossip.GossipMessage{
		Type: gossip.MsgBlock,
		From: node.ID,
		Data: []byte("block-123"),
	}
	gossip.Broadcast(node.PeerMgr.GetPeers(), msg)
}

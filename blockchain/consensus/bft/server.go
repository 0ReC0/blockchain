package bft

// запуск узла

import (
	"blockchain/network/gossip"
	"blockchain/network/peer"
	"blockchain/storage/blockchain"
	"blockchain/storage/txpool"
)

func StartNetwork(txPool *txpool.TransactionPool) {
	// Создаём блокчейн
	chain := blockchain.NewBlockchain()

	// Создаём узел с передачей всех необходимых аргументов
	node := NewNode("node1", ":3000", txPool, chain)

	// Добавляем пир
	node.PeerMgr.AddPeer(peer.NewPeer("peer1", ":3001", nil))

	// Запуск узла
	node.Start()

	// Запуск сетевых горутин
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

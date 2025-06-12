package bft

// точка входа узла

import (
	"blockchain/network/gossip"
	"blockchain/network/p2p"
	"blockchain/network/peer"
	"blockchain/storage/blockchain"
	"blockchain/storage/txpool"
	"crypto/tls"
	"fmt"
	"log"
	"time"
)

type Node struct {
	ID      string
	Addr    string
	PeerMgr *peer.PeerManager
	TxPool  *txpool.TransactionPool // Добавляем пул транзакций
	Chain   *blockchain.Blockchain  // Добавлено

}

func (n *Node) PerformHandshake(conn *tls.Conn) error {
	hs := p2p.NewHandshake(n.ID)
	data, _ := hs.Serialize()
	_, err := conn.Write(data)
	if err != nil {
		return err
	}

	// Read response
	buf := make([]byte, 1024)
	bytesRead, err := conn.Read(buf)
	remoteHS, err := p2p.DeserializeHandshake(buf[:bytesRead])
	if err != nil {
		return err
	}
	fmt.Printf("Handshake with %s successful\n", remoteHS.NodeID)
	return nil
}

func NewNode(id, addr string, txPool *txpool.TransactionPool, chain *blockchain.Blockchain) *Node {
	return &Node{
		ID:      id,
		Addr:    addr,
		PeerMgr: peer.NewPeerManager(),
		TxPool:  txPool,
		Chain:   chain,
	}
}

func (n *Node) Start() {
	fmt.Printf("Node %s started at %s\n", n.ID, n.Addr)
	go n.listenTLS()
}
func (n *Node) listenTLS() {
	config := p2p.GenerateTLSConfig()
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
	switch msg.Type {
	case gossip.StatePropose:
		block := n.CreateBlockFromPool() // Используем пул
		if block == nil {
			return
		}
		// Отправляем блок другим узлам
		n.BroadcastBlock(block)
	case gossip.MsgVote:
		// Обработка голосов
	}
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
func (n *Node) BroadcastBlock(block *blockchain.Block) {
	// Создаём сообщение с типом MsgBlock
	msg := &gossip.ConsensusMessage{
		Type:   gossip.MsgBlock,
		From:   n.ID,
		Height: block.Index,
		Round:  0, // Можно улучшить позже
		Block:  block,
	}

	data, err := msg.Encode()
	if err != nil {
		fmt.Printf("Failed to encode block message: %v\n", err)
		return
	}

	// Отправляем блок всем пеерам
	for _, peer := range n.PeerMgr.GetPeers() {
		_, err := peer.Connection.Write(data)
		if err != nil {
			fmt.Printf("Failed to send block to %s: %v\n", peer.ID, err)
		}
	}
}

func (n *Node) CreateBlockFromPool() *blockchain.Block {
	// 1. Получаем транзакции из пула
	txs := n.TxPool.GetTransactions(100)
	var validTxs []*txpool.Transaction

	// 2. Проверяем каждую транзакцию
	for _, tx := range txs {
		if !tx.Verify() {
			fmt.Printf("Invalid transaction: %s\n", tx.ID)
			continue
		}
		validTxs = append(validTxs, tx)
	}

	// 3. Если нет валидных транзакций — не создаём блок
	if len(validTxs) == 0 {
		fmt.Println("No valid transactions to propose")
		return nil
	}

	// 4. Получаем последний блок из цепочки
	latestBlock := n.Chain.GetLatestBlock()
	if latestBlock == nil {
		fmt.Println("Chain is empty or invalid")
		return nil
	}

	// 5. Создаём новый блок
	newBlock := &blockchain.Block{
		Index:        latestBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		PrevHash:     latestBlock.Hash,
		Transactions: validTxs,
		Validator:    n.ID,
	}

	// 6. Вычисляем хеш блока
	newBlock.Hash = newBlock.CalculateHash()

	return newBlock
}

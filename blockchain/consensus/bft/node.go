package bft

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

// Node представляет узел в BFT-сети
type Node struct {
	ID      string
	Addr    string
	PeerMgr *peer.PeerManager
	TxPool  *txpool.TransactionPool
	Chain   *blockchain.Blockchain
}

// NewNode создаёт новый узел
func NewNode(id string, addr string, txPool *txpool.TransactionPool, chain *blockchain.Blockchain) *Node {
	return &Node{
		ID:      id,
		Addr:    addr,
		TxPool:  txPool,
		Chain:   chain,
		PeerMgr: peer.NewPeerManager(),
	}
}

// Start запускает узел
func (n *Node) Start() {
	log.Printf("Node %s started at %s", n.ID, n.Addr)
	go n.listenTLS()
}

// listenTLS запускает TLS-сервер
func (n *Node) listenTLS() {
	config := p2p.GenerateClientTLSConfig()
	listener, err := tls.Listen("tcp", n.Addr, config)
	if err != nil {
		log.Fatalf("❌ Failed to start TLS listener: %v", err)
	}
	defer listener.Close()

	log.Printf("✅ TLS listener started on %s", n.Addr)

	for {
		rawConn, err := listener.Accept()
		if err != nil {
			log.Printf("❌ Failed to accept connection: %v", err)
			continue
		}

		tlsConn, ok := rawConn.(*tls.Conn)
		if !ok {
			log.Println("❌ Connection is not a TLS connection")
			rawConn.Close()
			continue
		}

		go n.handleSecureConnection(tlsConn)
	}
}

// handleSecureConnection обрабатывает безопасное соединение
func (n *Node) handleSecureConnection(conn *tls.Conn) {
	defer conn.Close()

	if err := n.PerformHandshake(conn); err != nil {
		log.Printf("❌ Handshake failed: %v", err)
		return
	}

	for {
		buf := make([]byte, 4096)
		nBytes, err := conn.Read(buf)
		if err != nil {
			log.Printf("❌ Connection closed: %v", err)
			return
		}

		msg, err := gossip.DecodeConsensusMessage(buf[:nBytes])
		if err != nil {
			log.Printf("❌ Failed to decode consensus message: %v", err)
			return
		}

		switch msg.Type {
		case gossip.MsgPing:
			n.handlePing(conn, msg)
		case gossip.MsgPong:
			log.Printf("🧾 Received pong from %s", msg.From)
		case gossip.StatePropose, gossip.StatePrevote, gossip.StatePrecommit, gossip.MsgBlock:
			go n.handleConsensusMessage(msg)
		default:
			log.Printf("🧾 Received unknown message type: %s from %s", msg.Type, msg.From)
		}
	}
}

// PerformHandshake выполняет рукопожатие с пиром
func (n *Node) PerformHandshake(conn *tls.Conn) error {
	hs := p2p.NewHandshake(n.ID)
	data, err := hs.Serialize()
	if err != nil {
		return fmt.Errorf("❌ failed to serialize handshake: %w", err)
	}

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("❌ failed to send handshake: %w", err)
	}

	// Чтение ответа
	buf := make([]byte, 1024)
	bytesRead, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("❌ failed to read handshake response: %w", err)
	}

	remoteHS, err := p2p.DeserializeHandshake(buf[:bytesRead])
	if err != nil {
		return fmt.Errorf("❌ failed to deserialize handshake: %w", err)
	}

	log.Printf("🤝 Handshake with %s successful", remoteHS.NodeID)
	return nil
}

// handlePing обрабатывает сообщение Ping
func (n *Node) handlePing(conn *tls.Conn, msg *gossip.ConsensusMessage) {
	log.Printf("🧾 Received ping from %s", msg.From)

	pong := &gossip.ConsensusMessage{
		Type:   gossip.MsgPong,
		From:   n.ID,
		Height: msg.Height,
		Round:  msg.Round,
	}

	data, err := pong.Encode()
	if err != nil {
		log.Printf("❌ Failed to encode pong message: %v", err)
		return
	}

	_, err = conn.Write(data)
	if err != nil {
		log.Printf("❌ Failed to send pong: %v", err)
	}
}

// handleConsensusMessage обрабатывает сообщения консенсуса
func (n *Node) handleConsensusMessage(msg *gossip.ConsensusMessage) {
	switch msg.Type {
	case gossip.StatePropose:
		block := n.CreateBlockFromPool()
		if block == nil {
			return
		}
		n.BroadcastBlock(block)
	case gossip.MsgBlock:
		log.Printf("🧾 Received block from %s", msg.From)
		// TODO: добавить валидацию и добавление блока в цепочку
	case gossip.MsgVote:
		log.Printf("🧾 Received vote from %s", msg.From)
		// TODO: обработать голос
	default:
		log.Printf("🧾 Unknown consensus message type: %s", msg.Type)
	}
}

// CreateBlockFromPool создаёт новый блок из пула транзакций
func (n *Node) CreateBlockFromPool() *blockchain.Block {
	txs := n.TxPool.GetTransactions(100)
	var validTxs []*txpool.Transaction

	for _, tx := range txs {
		if !tx.Verify() {
			log.Printf("🧾 Invalid transaction: %s", tx.ID)
			continue
		}
		validTxs = append(validTxs, tx)
	}

	if len(validTxs) == 0 {
		log.Println("🧾 No valid transactions to propose")
		return nil
	}

	latestBlock := n.Chain.GetLatestBlock()
	if latestBlock == nil {
		log.Println("🧾 Chain is empty or invalid")
		return nil
	}

	newBlock := &blockchain.Block{
		Index:        latestBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		PrevHash:     latestBlock.Hash,
		Transactions: validTxs,
		Validator:    n.ID,
	}

	newBlock.Hash = newBlock.CalculateHash()
	log.Printf("✅ Block %d created with %d transactions", newBlock.Index, len(validTxs))
	return newBlock
}

// BroadcastBlock рассылает блок всем пеерам
func (n *Node) BroadcastBlock(block *blockchain.Block) {
	msg := &gossip.ConsensusMessage{
		Type:   gossip.MsgBlock,
		From:   n.ID,
		Height: block.Index,
		Round:  0, // TODO: улучшить логику раундов
		Block:  block,
	}

	data, err := msg.Encode()
	if err != nil {
		log.Printf("❌ Failed to encode block message: %v", err)
		return
	}

	for _, peer := range n.PeerMgr.GetPeers() {
		_, err := peer.Connection.Write(data)
		if err != nil {
			log.Printf("❌ Failed to send block to %s: %v", peer.ID, err)
		}
	}
}

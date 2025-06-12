package bft

import (
	"blockchain/storage/blockchain"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// Message — тип сообщения между BFT-нодами
type TcpMessage struct {
	Type      string
	From      string
	Data      []byte
	Timestamp int64
}

// StartTCPServer — запускает TCP-сервер для BFT-нод
func StartTCPServer(bftNode *BFTNode) {
	listener, err := net.Listen("tcp", bftNode.Address)
	if err != nil {
		fmt.Printf("❌ Failed to start TCP server on %s: %v\n", bftNode.Address, err)
		return
	}
	defer listener.Close()

	fmt.Printf("📡 BFT Node listening on %s\n", bftNode.Address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("❌ Failed to accept connection: %v\n", err)
			continue
		}
		go handleConnection(conn, bftNode)
	}
}

// handleConnection — обработка входящих сообщений
func handleConnection(conn net.Conn, bftNode *BFTNode) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	var msg TcpMessage

	if err := decoder.Decode(&msg); err != nil {
		fmt.Printf("❌ Failed to decode message: %v\n", err)
		return
	}

	// Получаем текущий раунд (пример реализации — нужно адаптировать под вашу структуру)
	round := getCurrentRound(bftNode)

	switch msg.Type {
	case "proposal":
		block := &blockchain.Block{}
		if err := block.Deserialize(msg.Data); err != nil {
			fmt.Printf("❌ Failed to deserialize block: %v\n", err)
			return
		}
		round.ProposedBlock = msg.Data
		fmt.Printf("📬 Received proposal from %s\n", msg.From)

	case "prevote":
		round.Prevotes[msg.From] = msg.Data
		fmt.Printf("📬 Received prevote from %s\n", msg.From)

	case "precommit":
		round.Precommits[msg.From] = msg.Data
		fmt.Printf("📬 Received precommit from %s\n", msg.From)

	default:
		fmt.Printf("⚠️ Unknown message type: %s\n", msg.Type)
	}
}
func getCurrentRound(bftNode *BFTNode) *Round {
	// Пример — замените на вашу реальную логику
	return &Round{
		Height:      bftNode.Height,
		Round: bftNode.Round,
		Prevotes:    make(map[string][]byte),
		Precommits:  make(map[string][]byte),
	}
}

// BroadcastMessage — отправка сообщения всем пеерам
func BroadcastMessage(bftNode *BFTNode, msgType string, data []byte) {
	msg := TcpMessage{
		Type:      msgType,
		From:      bftNode.ID,
		Data:      data,
		Timestamp: time.Now().UnixNano(),
	}

	msgBytes, _ := json.Marshal(msg)

	for _, peer := range bftNode.Peers {
		go func(addr string) {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				fmt.Printf("❌ Can't connect to peer %s: %v\n", addr, err)
				return
			}
			defer conn.Close()

			_, err = conn.Write(msgBytes)
			if err != nil {
				fmt.Printf("❌ Failed to send message to %s: %v\n", addr, err)
			}
		}(peer)
	}
}

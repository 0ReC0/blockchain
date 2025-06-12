package bft

import (
	"blockchain/network/gossip"
	"blockchain/network/p2p"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// Message — тип сообщения между BFT-нодами
type TcpMessage struct {
	Type      gossip.MessageType
	From      string
	Data      []byte
	Timestamp int64
}

// StartTCPServer — запускает TCP-сервер для BFT-нод
func StartTCPServer(bftNode *BFTNode) {
	config := p2p.GenerateTLSConfig()
	listener, err := tls.Listen("tcp", bftNode.Address, config)
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

	// Оборачиваем в TLS
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		fmt.Println("❌ Connection is not a TLS connection")
		return
	}

	// Убедимся, что рукопожатие прошло
	if err := tlsConn.Handshake(); err != nil {
		fmt.Printf("❌ TLS handshake failed: %v\n", err)
		return
	}

	// Декодируем сообщение
	decoder := json.NewDecoder(tlsConn)
	var msg gossip.SignedConsensusMessage
	if err := decoder.Decode(&msg); err != nil {
		fmt.Printf("❌ Failed to decode message: %v\n", err)
		return
	}

	// Создаём хендлер
	handler := NewBFTMessageHandler(bftNode)

	// Обрабатываем сообщение
	fmt.Printf("📥 Received message from %s: %s\n", msg.From, msg.Type)
	handler.ProcessMessage(&msg)
}

// BroadcastMessage — отправка сообщения всем пеерам
func BroadcastMessage(bftNode *BFTNode, msgType gossip.MessageType, data []byte) {
	msg := TcpMessage{
		Type:      msgType,
		From:      bftNode.Address,
		Data:      data,
		Timestamp: time.Now().UnixNano(),
	}
	msgBytes, _ := json.Marshal(msg)

	for _, peer := range bftNode.Peers {
		go func(addr string) {
			// Используем TLS вместо обычного TCP
			conn, err := tls.Dial("tcp", addr, p2p.GenerateTLSConfig())
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

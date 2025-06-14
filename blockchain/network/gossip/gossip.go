package gossip

// протокол рассылки

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"blockchain/network/p2p"
	"blockchain/network/peer"
)

// SeenTransactionsSet — глобальный кэш хэшей обработанных транзакций
var SeenTransactionsSet = &SeenTransactions{
	seen: make(map[string]bool),
}

type SeenTransactions struct {
	seen map[string]bool
	mu   sync.RWMutex
}

// Add добавляет хэш транзакции в кэш
func (s *SeenTransactions) Add(hash string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seen[hash] = true
}

// Has проверяет, была ли транзакция с таким хэшем уже обработана
func (s *SeenTransactions) Has(hash string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.seen[hash]
}

// StartCleanup запускает периодическую очистку кэша
func (s *SeenTransactions) StartCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			<-ticker.C
			s.mu.Lock()
			s.seen = make(map[string]bool)
			s.mu.Unlock()
		}
	}()
}

func (m *GossipMessage) Encode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(m)
	return buf.Bytes(), err
}

func DecodeMessage(data []byte) (*GossipMessage, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	var msg GossipMessage
	err := dec.Decode(&msg)
	return &msg, err
}

func Broadcast(peers []*peer.Peer, msg *GossipMessage) error {
	for _, peer := range peers {
		conn, err := net.Dial("tcp", peer.Addr)
		if err != nil {
			fmt.Printf("Can't connect to peer %s: %v\n", peer.ID, err)
			continue
		}
		defer conn.Close()
		encoded, _ := msg.Encode()
		_, err = conn.Write(encoded)
		if err != nil {
			return err
		}
	}
	return nil
}

// BroadcastSignedConsensusMessage — рассылает подписанные сообщения всем пирам
func BroadcastSignedConsensusMessage(peers []*peer.Peer, msg *SignedConsensusMessage) error {
	txHash := fmt.Sprintf("%x", sha256.Sum256(msg.Data))
	if SeenTransactionsSet.Has(txHash) {
		return nil
	}

	// Помечаем как рассылаемую
	SeenTransactionsSet.Add(txHash)

	for _, peer := range peers {
		conn, err := tls.Dial("tcp", peer.Addr, p2p.GenerateClientTLSConfig())
		if err != nil {
			fmt.Printf("Can't connect to peer %s: %v\n", peer.ID, err)
			continue
		}
		defer conn.Close()
		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(msg); err != nil {
			return err
		}
	}
	return nil
}

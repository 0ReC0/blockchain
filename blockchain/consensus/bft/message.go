package bft

import (
	"blockchain/network/gossip"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// типы сообщений BFT

type Message struct {
	Type      gossip.MessageType
	Height    int64
	Round     int64
	Proposer  string
	Data      []byte
	Signature []byte
}

// SeenMessagesSet — кэш обработанных сообщений (по хэшу)
var SeenMessagesSet = &SeenMessages{
	seen: make(map[string]bool),
}

type SeenMessages struct {
	seen map[string]bool
	mu   sync.RWMutex
}

// Add добавляет хэш сообщения в кэш
func (s *SeenMessages) Add(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	s.seen[hash] = true
}

// Has проверяет, было ли сообщение уже обработано
func (s *SeenMessages) Has(data []byte) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	return s.seen[hash]
}

// StartCleanup запускает периодическую очистку кэша
func (s *SeenMessages) StartCleanup(interval time.Duration) {
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


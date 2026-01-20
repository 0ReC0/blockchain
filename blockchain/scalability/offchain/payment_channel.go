// blockchain/scalability/offchain/payment_channel.go
package offchain

import (
	"crypto/ecdsa"
	"errors"
	"time"

	"blockchain/crypto/signature"
)

// PaymentChannel представляет двусторонний платежный канал
type PaymentChannel struct {
	ID          string
	Participants [2]string       // Адреса участников
	Deposits    [2]float64      // Депозиты
	Nonce       int             // Счетчик транзакций
	StateHash   string          // Хэш текущего состояния
	Timeout     time.Time       // Время жизни канала
	Settlement  *Settlement     // Итоговое распределение
	Signatures  [2][]byte       // Подписи участников
	PublicKeys  [2]*ecdsa.PublicKey
}

// Settlement — итоговое состояние канала
type Settlement struct {
	Amounts [2]float64
	Nonce   int
}

// ChannelManager — управление каналами
type ChannelManager struct {
	channels map[string]*PaymentChannel
}

// NewChannelManager — создает новый ChannelManager
func NewChannelManager() *ChannelManager {
	return &ChannelManager{
		channels: make(map[string]*PaymentChannel),
	}
}

// GetChannel returns a channel by its ID
func (cm *ChannelManager) GetChannel(id string) (*PaymentChannel, bool) {
	channel, exists := cm.channels[id]
	return channel, exists
}

// AddChannel adds a channel to the manager
func (cm *ChannelManager) AddChannel(channel *PaymentChannel) {
	cm.channels[channel.ID] = channel
}

// CreateChannel — создает новый канал
func (cm *ChannelManager) CreateChannel(id string, a, b string, depositA, depositB float64, pubKeyA, pubKeyB *ecdsa.PublicKey, timeout time.Time) (*PaymentChannel, error) {
	if depositA <= 0 || depositB <= 0 {
		return nil, errors.New("deposits must be positive")
	}
	channel := &PaymentChannel{
		ID: id,
		Participants: [2]string{a, b},
		Deposits:     [2]float64{depositA, depositB},
		Nonce:        0,
		StateHash:    "",
		Timeout:      timeout,
		PublicKeys:   [2]*ecdsa.PublicKey{pubKeyA, pubKeyB},
	}
	cm.channels[id] = channel
	return channel, nil
}

// UpdateState — обновляет состояние канала
func (pc *PaymentChannel) UpdateState(amountA, amountB float64, nonce int, sigA, sigB []byte) error {
	if nonce != pc.Nonce+1 {
		return errors.New("invalid nonce")
	}
	if amountA+amountB != pc.Deposits[0]+pc.Deposits[1] {
		return errors.New("invalid amounts")
	}
	if !signature.Verify(pc.PublicKeys[0], []byte(pc.ID), sigA) {
		return errors.New("invalid signature from participant A")
	}
	if !signature.Verify(pc.PublicKeys[1], []byte(pc.ID), sigB) {
		return errors.New("invalid signature from participant B")
	}
	pc.Settlement = &Settlement{
		Amounts: [2]float64{amountA, amountB},
		Nonce:   nonce,
	}
	pc.Nonce = nonce
	return nil
}

// Finalize — завершает канал и возвращает итоговое состояние
func (pc *PaymentChannel) Finalize() (*Settlement, error) {
	if pc.Settlement == nil {
		return nil, errors.New("no settlement provided")
	}
	if time.Now().After(pc.Timeout) {
		return nil, errors.New("channel timeout expired")
	}
	return pc.Settlement, nil
}

// Close — закрывает канал
func (pc *PaymentChannel) Close(settlement *Settlement) error {
	if settlement.Nonce != pc.Nonce {
		return errors.New("invalid settlement nonce")
	}
	pc.Settlement = settlement
	return nil
}

package bft

import (
	"crypto/ecdsa"
	"crypto/tls"
	"log"

	"blockchain/crypto/signature"
	"blockchain/network/gossip"
	"blockchain/network/p2p"
	"blockchain/storage/blockchain"
)

// Константы для логов
const (
	ErrNilData           = "❌ Received message with nil data"
	ErrDeserialize       = "❌ Failed to deserialize block"
	ErrNilCurrentRound   = "❌ CurrentRound is nil — cannot validate proposer"
	ErrWrongProposer     = "❌ Block proposer %s is not current proposer %s"
	ErrValidatorMismatch = "❌ Validator mismatch: msg.From=%s, block.Validator=%s"
	ErrNoPubKey          = "❌ Validator has no public key: %v"
	ErrInvalidSig        = "❌ Invalid signature from %s"
)

// BFTMessageHandler обрабатывает входящие сообщения BFT протокола
type BFTMessageHandler struct {
	Node *BFTNode
}

// NewBFTMessageHandler создаёт новый обработчик сообщений
func NewBFTMessageHandler(node *BFTNode) *BFTMessageHandler {
	return &BFTMessageHandler{Node: node}
}

// ProcessMessage обрабатывает сообщение в зависимости от его типа
func (h *BFTMessageHandler) ProcessMessage(msg *gossip.SignedConsensusMessage) {
	switch msg.Type {
	case gossip.StatePropose:
		h.HandlePropose(msg)
	case gossip.StatePrevote:
		h.HandlePrevote(msg)
	case gossip.StatePrecommit:
		h.HandlePrecommit(msg)
	case gossip.MsgStatus:
		h.handleStatusMessage(msg)
	default:
		h.HandleUnknown(msg)
	}
}

// handleStatusMessage обрабатывает запросы статуса
func (h *BFTMessageHandler) handleStatusMessage(msg *gossip.SignedConsensusMessage) {
	if msg.From == h.Node.Address {
		return
	}

	latestBlock := h.Node.Chain.GetLatestBlock()
	if latestBlock == nil {
		log.Println("❌ Chain is empty")
		return
	}

	response := &gossip.ConsensusMessage{
		Type:   gossip.MsgStatus,
		From:   h.Node.Address,
		Height: latestBlock.Index,
	}

	data, err := response.Encode()
	if err != nil {
		log.Printf("❌ Failed to encode status response: %v\n", err)
		return
	}

	conn, err := tls.Dial("tcp", msg.From, p2p.GenerateClientTLSConfig())
	if err != nil {
		log.Printf("❌ Failed to connect to %s: %v\n", msg.From, err)
		return
	}
	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		log.Printf("❌ Failed to send status response to %s: %v\n", msg.From, err)
	}
}

// verifySignature проверяет подпись сообщения
func (h *BFTMessageHandler) verifySignature(msg *gossip.SignedConsensusMessage, pubKey *ecdsa.PublicKey) bool {
	if !ecdsa.VerifyASN1(pubKey, msg.Data, msg.Signature) {
		log.Printf(ErrInvalidSig, msg.From)
		return false
	}
	return true
}

// getPublicKey возвращает публичный ключ по адресу
func (h *BFTMessageHandler) getPublicKey(address string) (*ecdsa.PublicKey, error) {
	pubKey, err := signature.GetPublicKey(address)
	if err != nil {
		log.Printf(ErrNoPubKey, address)
		return nil, err
	}
	return pubKey, nil
}

// HandlePropose обрабатывает сообщение с предложением блока
func (h *BFTMessageHandler) HandlePropose(msg *gossip.SignedConsensusMessage) {
	if SeenMessagesSet.Has(msg.Data) {
		log.Println("❌ Duplicate propose message ignored")
		return
	}
	SeenMessagesSet.Add(msg.Data)

	if msg.Data == nil {
		log.Println(ErrNilData)
		return
	}

	block := &blockchain.Block{}
	if err := block.Deserialize(msg.Data); err != nil {
		log.Printf("%s: %v\n", ErrDeserialize, err)
		return
	}

	if h.Node.CurrentRound == nil {
		log.Println(ErrNilCurrentRound)
		return
	}

	if msg.From != h.Node.CurrentRound.Proposer {
		log.Printf(ErrWrongProposer, msg.From, h.Node.CurrentRound.Proposer)
		return
	}

	if msg.From != block.Validator {
		log.Printf(ErrValidatorMismatch, msg.From, block.Validator)
		return
	}

	pubKey, err := h.getPublicKey(block.Validator)
	if err != nil {
		return
	}

	if !h.verifySignature(msg, pubKey) {
		return
	}

	h.Node.CurrentRound.ProposedBlock = msg.Data
	log.Printf("✅ Block signature verified successfully, hash: %x\n", block.Hash)
}

// HandlePrevote обрабатывает сообщение prevote
func (h *BFTMessageHandler) HandlePrevote(msg *gossip.SignedConsensusMessage) {
	log.Printf("[DEBUG] Received prevote from %s, round: %d, height: %d\n", msg.From, msg.Round, msg.Height)

	if !h.Node.IsValidator(msg.From) {
		log.Printf("❌ Ignoring prevote from non-validator: %s\n", msg.From)
		return
	}

	if msg.Round != h.Node.Round || msg.Height != h.Node.Height {
		log.Printf("❌ Ignoring prevote for wrong round/height: %d/%d (expected %d/%d)\n",
			msg.Round, msg.Height, h.Node.Round, h.Node.Height)
		return
	}

	if SeenMessagesSet.Has(msg.Data) {
		return
	}
	SeenMessagesSet.Add(msg.Data)

	pubKey, err := h.getPublicKey(msg.From)
	if err != nil {
		return
	}

	if !h.verifySignature(msg, pubKey) {
		return
	}

	h.Node.CurrentRound.Prevotes[msg.From] = msg.Signature
	log.Printf("🗳 Prevote from %s verified successfully\n", msg.From)
}

// HandlePrecommit обрабатывает сообщение precommit
func (h *BFTMessageHandler) HandlePrecommit(msg *gossip.SignedConsensusMessage) {
	log.Printf("[DEBUG] Received precommit from %s, round: %d, height: %d\n", msg.From, msg.Round, msg.Height)

	if !h.Node.IsValidator(msg.From) {
		log.Printf("❌ Ignoring precommit from non-validator: %s\n", msg.From)
		return
	}

	if msg.Round != h.Node.Round || msg.Height != h.Node.Height {
		log.Printf("❌ Ignoring precommit for wrong round/height: %d/%d (expected %d/%d)\n",
			msg.Round, msg.Height, h.Node.Round, h.Node.Height)
		return
	}

	pubKey, err := h.getPublicKey(msg.From)
	if err != nil {
		return
	}

	if !h.verifySignature(msg, pubKey) {
		return
	}

	h.Node.CurrentRound.Precommits[msg.From] = msg.Data
	log.Printf("✅ Precommit received from %s\n", msg.From)
}

// HandleUnknown обрабатывает неизвестные типы сообщений
func (h *BFTMessageHandler) HandleUnknown(msg *gossip.SignedConsensusMessage) {
	log.Printf("❌ Unknown message type: %s\n", msg.Type)
}

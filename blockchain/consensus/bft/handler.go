package bft

import (
	"blockchain/crypto/signature"
	"blockchain/network/gossip"
	"blockchain/storage/blockchain"
	"crypto/ecdsa"
	"fmt"
)

type BFTMessageHandler struct {
	Node *BFTNode
}

func NewBFTMessageHandler(node *BFTNode) *BFTMessageHandler {
	return &BFTMessageHandler{Node: node}
}

func (h *BFTMessageHandler) ProcessMessage(msg *gossip.SignedConsensusMessage) {
	switch msg.Type {
	case gossip.StatePropose:
		h.HandlePropose(msg)
	case gossip.StatePrevote:
		h.HandlePrevote(msg)
	case gossip.StatePrecommit:
		h.HandlePrecommit(msg)
	default:
		h.HandleUnknown(msg)
	}
}

func (h *BFTMessageHandler) HandlePropose(msg *gossip.SignedConsensusMessage) {
	// Проверяем, не было ли это сообщение уже обработано
	if SeenMessagesSet.Has(msg.Data) {
		fmt.Println("❌ Duplicate propose message ignored")
		return
	}
	SeenMessagesSet.Add(msg.Data)

	// Проверяем, что данные не nil
	if msg.Data == nil {
		fmt.Println("❌ Received propose message with nil data")
		return
	}

	block := &blockchain.Block{}
	if err := block.Deserialize(msg.Data); err != nil {
		fmt.Printf("❌ Failed to deserialize block: %v\n", err)
		return
	}
	fmt.Printf("📬 Received proposal from %s, block hash: %x\n", msg.From, block.Hash)

	// Остальная логика
	fmt.Printf("🧾 Validator in block: %s\n", block.Validator)
	fmt.Printf("🧾 Validator in message: %s\n", msg.From)
	if h.Node.CurrentRound == nil {
		fmt.Println("❌ CurrentRound is nil — cannot validate proposer")
		return
	}

	if msg.From != h.Node.CurrentRound.Proposer {
		fmt.Printf("❌ Block proposer %s is not current proposer %s\n", msg.From, h.Node.CurrentRound.Proposer)
		return
	}

	if msg.From != block.Validator {
		fmt.Printf("❌ Validator mismatch: msg.From=%s, block.Validator=%s\n", msg.From, block.Validator)
		return
	}

	// Получаем публичный ключ
	pubKey, err := signature.GetPublicKey(block.Validator)
	if err != nil {
		fmt.Printf("❌ Validator has no public key: %v\n", err)
		return
	}

	if !ecdsa.VerifyASN1(pubKey, msg.Data, msg.Signature) {
		fmt.Printf("❌ Invalid proposal signature from %s\n", msg.From)
		return
	}

	// Сохраняем блок
	h.Node.CurrentRound.ProposedBlock = msg.Data
	fmt.Println("✅ Block signature verified successfully")
}

func (h *BFTMessageHandler) HandlePrevote(msg *gossip.SignedConsensusMessage) {
	fmt.Printf("[DEBUG] Received prevote from %s, round: %d, height: %d\n", msg.From, msg.Round, msg.Height)
	if !h.Node.IsValidator(msg.From) {
		fmt.Printf("❌ Ignoring prevote from non-validator: %s\n", msg.From)
		return
	}
	if msg.Round != h.Node.Round || msg.Height != h.Node.Height {
		fmt.Printf("❌ Ignoring prevote for wrong round/height: %d/%d (expected %d/%d)\n", msg.Round, msg.Height, h.Node.Round, h.Node.Height)
		return
	}
	if SeenMessagesSet.Has(msg.Data) {
		return
	}
	SeenMessagesSet.Add(msg.Data)

	pubKey, err := signature.GetPublicKey(msg.From)
	if err != nil {
		fmt.Printf("❌ Failed to get public key for %s\n", msg.From)
		return
	}

	// Проверяем подпись с помощью ecdsa.VerifyASN1
	if !ecdsa.VerifyASN1(pubKey, msg.Data, msg.Signature) {
		fmt.Printf("❌ Invalid prevote signature from %s\n", msg.From)
		return
	}

	h.Node.CurrentRound.Prevotes[msg.From] = msg.Signature
	fmt.Printf("🗳 Prevote from %s verified successfully\n", msg.From)
}

func (h *BFTMessageHandler) HandlePrecommit(msg *gossip.SignedConsensusMessage) {
	fmt.Printf("[DEBUG] Received precommit from %s, round: %d, height: %d\n", msg.From, msg.Round, msg.Height)
	if !h.Node.IsValidator(msg.From) {
		fmt.Printf("❌ Ignoring precommit from non-validator: %s\n", msg.From)
		return
	}
	if msg.Round != h.Node.Round || msg.Height != h.Node.Height {
		fmt.Printf("❌ Ignoring precommit for wrong round/height: %d/%d (expected %d/%d)\n", msg.Round, msg.Height, h.Node.Round, h.Node.Height)
		return
	}
	pubKey, err := signature.GetPublicKey(msg.From)
	if err != nil {
		fmt.Printf("❌ Failed to get public key for %s\n", msg.From)
		return
	}
	if !ecdsa.VerifyASN1(pubKey, msg.Data, msg.Signature) {
		fmt.Printf("❌ Invalid precommit signature from %s\n", msg.From)
		return
	}
	h.Node.CurrentRound.Precommits[msg.From] = msg.Data
	fmt.Printf("✅ Precommit received from %s\n", msg.From)
}

func (h *BFTMessageHandler) HandleUnknown(msg *gossip.SignedConsensusMessage) {
	// fmt.Printf("Unknown message type: %s\n", msg.Type)
}

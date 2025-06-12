package bft

import (
	"blockchain/crypto/signature"
	"blockchain/network/gossip"
	"blockchain/storage/blockchain"
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
	block := &blockchain.Block{}
	if err := block.Deserialize(msg.Data); err != nil {
		return
	}

	// Получаем публичный ключ
	pubKey, err := signature.GetPublicKey(block.Validator)
	if err != nil {
		return
	}

	// Проверяем подпись без поля Signature
	if !signature.Verify(pubKey, block.SerializeWithoutSignature(), block.Signature) {
		fmt.Println("❌ Invalid block signature")
		return
	}

	// Сохраняем блок
	h.Node.CurrentRound.ProposedBlock = msg.Data
}

func (h *BFTMessageHandler) HandlePrevote(msg *gossip.SignedConsensusMessage) {
	h.Node.CurrentRound.Prevotes[msg.From] = msg.Data
}

func (h *BFTMessageHandler) HandlePrecommit(msg *gossip.SignedConsensusMessage) {
	h.Node.CurrentRound.Precommits[msg.From] = msg.Data
}

func (h *BFTMessageHandler) HandleUnknown(msg *gossip.SignedConsensusMessage) {
	// fmt.Printf("Unknown message type: %s\n", msg.Type)
}

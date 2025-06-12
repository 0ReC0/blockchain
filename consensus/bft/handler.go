// bft/handler.go

package bft

import (
	"blockchain/network/gossip"
	"blockchain/network/peer"
	"fmt"
)

type BFTMessageHandler struct {
	PeerManager *peer.PeerManager
}

func NewBFTMessageHandler(peerMgr *peer.PeerManager) *BFTMessageHandler {
	return &BFTMessageHandler{
		PeerManager: peerMgr,
	}
}

func (h *BFTMessageHandler) ProcessMessage(msg *gossip.ConsensusMessage) {
	switch msg.Type {
	case gossip.MsgPropose:
		h.HandlePropose(msg)
	case gossip.MsgPrevote:
		h.HandlePrevote(msg)
	case gossip.MsgPrecommit:
		h.HandlePrecommit(msg)
	default:
		fmt.Printf("Unknown message type: %s\n", msg.Type)
	}
}

func (h *BFTMessageHandler) HandlePropose(msg *gossip.ConsensusMessage) {
	fmt.Printf("[BFT] Received Propose: %s (Height: %d, Round: %d)\n", msg.From, msg.Height, msg.Round)
	// Здесь можно проверить блок, сохранить и отправить prevote
}

func (h *BFTMessageHandler) HandlePrevote(msg *gossip.ConsensusMessage) {
	fmt.Printf("[BFT] Received Prevote: %s (Height: %d, Round: %d)\n", msg.From, msg.Height, msg.Round)
	// Здесь можно собрать prevotes и перейти к precommit
}

func (h *BFTMessageHandler) HandlePrecommit(msg *gossip.ConsensusMessage) {
	fmt.Printf("[BFT] Received Precommit: %s (Height: %d, Round: %d)\n", msg.From, msg.Height, msg.Round)
	// Здесь можно проверить, собрано ли 2/3+ precommit'ов
}

func (h *BFTMessageHandler) HandleCommit(msg *gossip.ConsensusMessage) {
	fmt.Printf("[BFT] Received Commit: %s (Height: %d, Round: %d)\n", msg.From, msg.Height, msg.Round)
	// Здесь можно добавить блок в блокчейн
}

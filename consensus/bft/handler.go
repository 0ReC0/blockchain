// bft/handler.go

package bft

import (
	"../../network/gossip"
	"../../network/peer"
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
	// Пример базовой обработки
	switch msg.Type {
	case gossip.MsgPropose:
		// TODO: обработать MsgPropose
	case gossip.MsgPrevote:
		// TODO: обработать MsgPrevote
	case gossip.MsgPrecommit:
		// TODO: обработать MsgPrecommit
	default:
		// TODO: неизвестное сообщение
	}
}

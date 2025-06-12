// bft/handler.go

package bft

import (
	"blockchain/network/peer"
)

type BFTMessageHandler struct {
	PeerManager *peer.PeerManager
}

func NewBFTMessageHandler(peerMgr *peer.PeerManager) *BFTMessageHandler {
	return &BFTMessageHandler{
		PeerManager: peerMgr,
	}
}

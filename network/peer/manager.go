package peer

import (
	"fmt"

	"../../security/sybil"
)

// управление пирингом

type PeerManager struct {
	peers map[string]*Peer
}

func NewPeerManager() *PeerManager {
	return &PeerManager{
		peers: make(map[string]*Peer),
	}
}

var sybilGuard *sybil.SybilGuard

func init() {
	validators := []string{"validator1", "validator2"}
	sybilGuard = sybil.NewSybilGuard(validators)
}

func (pm *PeerManager) AddPeer(p *Peer) {
	if !sybilGuard.RegisterNode(p.ID) {
		fmt.Printf("Failed to register peer %s: Sybil check failed\n", p.ID)
		return
	}
	pm.peers[p.ID] = p
	fmt.Printf("Peer added: %s\n", p.ID)
}

func (pm *PeerManager) RemovePeer(id string) {
	delete(pm.peers, id)
	fmt.Printf("Peer removed: %s\n", id)
}

func (pm *PeerManager) GetPeers() []*Peer {
	var list []*Peer
	for _, p := range pm.peers {
		list = append(list, p)
	}
	return list
}

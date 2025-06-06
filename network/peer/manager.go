package peer

import "fmt"

type PeerManager struct {
	peers map[string]*Peer
}

func NewPeerManager() *PeerManager {
	return &PeerManager{
		peers: make(map[string]*Peer),
	}
}

func (pm *PeerManager) AddPeer(p *Peer) {
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

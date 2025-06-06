package peer

type Peer struct {
	ID   string
	Addr string
}

func NewPeer(id, addr string) *Peer {
	return &Peer{ID: id, Addr: addr}
}

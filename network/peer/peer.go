package peer

import "crypto/tls"

// модель узла

type Peer struct {
	ID         string
	Addr       string
	Connection *tls.Conn // Добавляем поле

}

func NewPeer(id, addr string, conn ...*tls.Conn) *Peer {
	var connection *tls.Conn
	if len(conn) > 0 {
		connection = conn[0]
	}
	return &Peer{
		ID:         id,
		Addr:       addr,
		Connection: connection,
	}
}

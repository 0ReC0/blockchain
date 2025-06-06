package multiaddr

import (
	"fmt"
	"strings"
)

type Multiaddr struct {
	proto string
	addr  string
	port  string
}

func NewMultiaddr(proto, addr, port string) *Multiaddr {
	return &Multiaddr{
		proto: proto,
		addr:  addr,
		port:  port,
	}
}

func (m *Multiaddr) String() string {
	return fmt.Sprintf("/%s/%s/%s", m.proto, m.addr, m.port)
}

func ParseMultiaddr(s string) (*Multiaddr, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid multiaddr format")
	}
	return &Multiaddr{
		proto: parts[1],
		addr:  parts[2],
		port:  parts[3],
	}, nil
}

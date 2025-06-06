package gossip

// протокол рассылки

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
)

func (m *Message) Encode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(m)
	return buf.Bytes(), err
}

func DecodeMessage(data []byte) (*Message, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	var msg Message
	err := dec.Decode(&msg)
	return &msg, err
}

func Broadcast(peers []*Peer, msg *Message) error {
	for _, peer := range peers {
		conn, err := net.Dial("tcp", peer.Addr)
		if err != nil {
			fmt.Printf("Can't connect to peer %s: %v\n", peer.ID, err)
			continue
		}
		defer conn.Close()
		encoded, _ := msg.Encode()
		_, err = conn.Write(encoded)
		if err != nil {
			return err
		}
	}
	return nil
}

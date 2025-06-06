package peer

import (
	"fmt"
	"net"
	"time"
)

const BroadcastAddr = "255.255.255.255:30000"

func BroadcastPresence(addr string) {
	conn, _ := net.Dial("udp", BroadcastAddr)
	defer conn.Close()

	msg := fmt.Sprintf("NODE_PRESENT:%s", addr)
	for {
		conn.Write([]byte(msg))
		time.Sleep(5 * time.Second)
	}
}

func ListenForPeers() {
	listener, _ := net.ListenUDP("udp", &net.UDPAddr{Port: 30000})
	buf := make([]byte, 1024)

	for {
		n, _, _ := listener.ReadFromUDP(buf)
		msg := string(buf[:n])
		if len(msg) > 11 && msg[:11] == "NODE_PRESENT" {
			addr := msg[12:]
			fmt.Printf("Discovered peer: %s\n", addr)
		}
	}
}

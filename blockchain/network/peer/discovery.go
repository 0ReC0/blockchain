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
	if udpConn == nil {
		panic("❌ UDP socket not initialized")
	}

	buf := make([]byte, 1024)
	for {
		n, addr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("❌ UDP read error: %v\n", err)
			continue
		}
		fmt.Printf("📬 Received from %s: %s\n", addr, string(buf[:n]))
	}
}

var udpConn *net.UDPConn

// InitUDPSocket инициализирует UDP-сокет на указанном порту
func InitUDPSocket(port string) {
	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to resolve UDP address: %v", err))
	}

	udpConn, err = net.ListenUDP("udp", addr)
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to start UDP listener: %v", err))
	}

	fmt.Printf("📡 UDP listener started on %s\n", port)
}

package main

import (
	p2p "./network/p2p"
	rpc "./network/rpc"
)

func main() {
	go p2p.StartNetwork()
	rpc.StartRPCServer(":8080")
	select {}
}

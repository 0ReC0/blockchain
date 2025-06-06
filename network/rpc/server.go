package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type RPCService struct{}

func (s *RPCService) HandleTx(req []byte, reply *string) error {
	fmt.Printf("Received transaction: %s\n", req)
	*reply = "OK"
	return nil
}

func (s *RPCService) HandleBlock(req []byte, reply *string) error {
	fmt.Printf("Received block: %s\n", req)
	*reply = "OK"
	return nil
}

func StartRPCServer(addr string) {
	rpc := new(RPCService)
	http.Handle("/tx", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tx json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		var reply string
		rpc.HandleTx(tx, &reply)
		w.Write([]byte(reply))
	}))

	http.ListenAndServe(addr, nil)
}

package api

import (
	"encoding/json"
	"net/http"

	"../../storage/blockchain"
	"../../storage/txpool"
)

type APIServer struct {
	Chain  *blockchain.Blockchain
	TxPool *txpool.TransactionPool
}

func NewAPIServer(chain *blockchain.Blockchain, txPool *txpool.TransactionPool) *APIServer {
	return &APIServer{
		Chain:  chain,
		TxPool: txPool,
	}
}

func (s *APIServer) Start(addr string) error {
	http.HandleFunc("/blocks", s.handleBlocks)
	http.HandleFunc("/transactions", s.handleTransactions)
	return http.ListenAndServe(addr, nil)
}

func (s *APIServer) handleBlocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.Chain.Blocks)
}

func (s *APIServer) handleTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.TxPool.GetTransactions(100))
}

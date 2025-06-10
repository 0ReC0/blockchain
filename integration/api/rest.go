package api

import (
	"encoding/json"
	"net/http"

	"../../contracts/execution"
	"../../storage/blockchain"
	"../../storage/txpool"
)

type APIServer struct {
	Chain           *blockchain.Blockchain
	TxPool          *txpool.TransactionPool
	ContractHandler *execution.ContractHandler
}

func NewAPIServer(chain *blockchain.Blockchain, txPool *txpool.TransactionPool) *APIServer {
	return &APIServer{
		Chain:           chain,
		TxPool:          txPool,
		ContractHandler: execution.NewContractHandler(),
	}
}

func (s *APIServer) Start(addr string) error {
	http.HandleFunc("/blocks", s.handleBlocks)
	http.HandleFunc("/transactions", s.handleTransactions)
	http.HandleFunc("/contract/deploy", s.handleDeployContract)
	http.HandleFunc("/contract/call", s.handleCallContract)
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

func (s *APIServer) handleCallContract(w http.ResponseWriter, r *http.Request) {
	type CallRequest struct {
		Address string        `json:"address"`
		Method  string        `json:"method"`
		Args    []interface{} `json:"args"`
	}

	var req CallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Вызываем контракт
	result, err := s.ContractHandler.CallERC20(req.Address, req.Method, req.Args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"result": result,
	})
}
func (s *APIServer) handleDeployContract(w http.ResponseWriter, r *http.Request) {
	type DeployRequest struct {
		Name     string `json:"name"`
		Symbol   string `json:"symbol"`
		Decimals int    `json:"decimals"`
		Supply   uint64 `json:"supply"`
	}

	var req DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Деплоим токен
	addr := s.ContractHandler.DeployERC20(req.Name, req.Symbol, req.Decimals, req.Supply)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"address": addr,
	})
}

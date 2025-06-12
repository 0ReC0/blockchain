package api

import (
	"encoding/json"
	"net/http"

	"blockchain/contracts/execution"
	"blockchain/storage/blockchain"
	"blockchain/storage/txpool"
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
	http.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			s.handleTransactions(w, r)
		case "POST":
			s.handleAddTransaction(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
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

func (s *APIServer) handleAddTransaction(w http.ResponseWriter, r *http.Request) {
	tx := &txpool.Transaction{}
	if err := json.NewDecoder(r.Body).Decode(tx); err != nil {
		http.Error(w, "Invalid transaction format", http.StatusBadRequest)
		return
	}

	if !tx.Verify() {
		http.Error(w, "Invalid transaction signature", http.StatusBadRequest)
		return
	}

	s.TxPool.AddTransaction(tx)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":      "success",
		"transaction": tx.ID,
	})
}

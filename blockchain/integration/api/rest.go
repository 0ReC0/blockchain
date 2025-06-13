// api/rest.go
package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"blockchain/crypto/signature"
	"blockchain/security/audit"
	"blockchain/storage/blockchain"
	"blockchain/storage/txpool"

	"github.com/dgraph-io/badger/v4"
)

type APIServer struct {
	Chain  *blockchain.Blockchain
	TxPool *txpool.TransactionPool
}

var auditor *audit.SecurityAuditor

func NewAPIServer(chain *blockchain.Blockchain, txPool *txpool.TransactionPool, auditorInstance *audit.SecurityAuditor) *APIServer {
	auditor = auditorInstance

	if txPool == nil {
		txPool = txpool.NewTransactionPool()
	}

	return &APIServer{
		Chain:  chain,
		TxPool: txPool,
	}
}

func (s *APIServer) Start(addr string) error {
	http.HandleFunc("/audit", enableCORS(s.handleSecurityAudit))
	http.HandleFunc("/blocks", enableCORS(s.handleBlocks))
	http.HandleFunc("/register", s.handleRegisterPublicKey)
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
	return http.ListenAndServe(addr, nil)
}

// ===== Блоки =====

func (s *APIServer) handleBlocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.Chain == nil {
		http.Error(w, "Blockchain not initialized", http.StatusInternalServerError)
		return
	}

	db := s.Chain.DB()
	if db == nil {
		http.Error(w, "Database not available", http.StatusInternalServerError)
		return
	}

	var blocks []*blockchain.Block = make([]*blockchain.Block, 0) // ✅ Инициализируем срез
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			val, _ := item.ValueCopy(nil)
			block := &blockchain.Block{}
			block.Deserialize(val)
			if block.Transactions == nil {
				block.Transactions = []*txpool.Transaction{}
			}
			blocks = append(blocks, block)
		}
		return nil
	})

	if err != nil {
		http.Error(w, "Failed to read blocks from DB", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(blocks)
}

// ===== Транзакции =====

func (s *APIServer) handleTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.TxPool.GetTransactions(100))
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

// ===== Аудит =====

func (s *APIServer) handleSecurityAudit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(auditor.GetEvents())
}

// ===== Регистрация ключей =====

func (s *APIServer) handleRegisterPublicKey(w http.ResponseWriter, r *http.Request) {
	type RegisterRequest struct {
		Address string `json:"address"`
		PubKey  string `json:"pubKey"`
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	pubKeyBytes, err := hex.DecodeString(req.PubKey)
	if err != nil {
		http.Error(w, "Invalid public key format", http.StatusBadRequest)
		return
	}

	pubKey, err := signature.ParsePublicKey(pubKeyBytes)
	if err != nil {
		http.Error(w, "Failed to parse public key", http.StatusBadRequest)
		return
	}

	// Сохраняем в реестр
	signature.RegisterPublicKey(req.Address, pubKey)

	fmt.Printf("🔑 Public key registered for %s\n", req.Address)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Public key registered for %s", req.Address)
}

// ===== CORS =====

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

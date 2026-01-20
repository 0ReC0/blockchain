// blockchain/integration/api/offchain.go
package api

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"blockchain/crypto/signature"
	"blockchain/scalability/offchain"
)

var channelManager = offchain.NewChannelManager()

// handleCreateChannel — создает новый платежный канал
func (s *APIServer) handleCreateChannel(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		ID       string  `json:"id"`
		AddrA    string  `json:"addr_a"`
		AddrB    string  `json:"addr_b"`
		DepositA float64 `json:"deposit_a"`
		DepositB float64 `json:"deposit_b"`
		PubKeyA  string  `json:"pubkey_a"`
		PubKeyB  string  `json:"pubkey_b"`
		Timeout  int64   `json:"timeout"` // Unix timestamp
	}
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	pubKeyA, err := signature.LoadPublicKey(req.PubKeyA)
	if err != nil {
		http.Error(w, "Invalid public key A", http.StatusBadRequest)
		return
	}
	pubKeyB, err := signature.LoadPublicKey(req.PubKeyB)
	if err != nil {
		http.Error(w, "Invalid public key B", http.StatusBadRequest)
		return
	}
	_, err = channelManager.CreateChannel(
		req.ID, req.AddrA, req.AddrB, req.DepositA, req.DepositB,
		pubKeyA, pubKeyB, time.Unix(req.Timeout, 0),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Channel created", "id": req.ID})
}

// handleUpdateChannel — обновляет состояние канала
func (s *APIServer) handleUpdateChannel(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		ID       string  `json:"id"`
		AmountA  float64 `json:"amount_a"`
		AmountB  float64 `json:"amount_b"`
		Nonce    int     `json:"nonce"`
		SigA     string  `json:"sig_a"`
		SigB     string  `json:"sig_b"`
	}
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	sigA, err := hex.DecodeString(req.SigA)
	if err != nil {
		http.Error(w, "Invalid signature A", http.StatusBadRequest)
		return
	}
	sigB, err := hex.DecodeString(req.SigB)
	if err != nil {
		http.Error(w, "Invalid signature B", http.StatusBadRequest)
		return
	}
	channel, exists := channelManager.GetChannel(req.ID)
	if !exists {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}
	if err := channel.UpdateState(req.AmountA, req.AmountB, req.Nonce, sigA, sigB); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Channel updated", "id": req.ID})
}

// handleFinalizeChannel — завершает канал
func (s *APIServer) handleFinalizeChannel(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		ID string `json:"id"`
	}
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	channel, exists := channelManager.GetChannel(req.ID)
	if !exists {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}
	settlement, err := channel.Finalize()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "Channel finalized",
		"id":     req.ID,
		"settlement": map[string]interface{}{
			"amount_a": settlement.Amounts[0],
			"amount_b": settlement.Amounts[1],
			"nonce":    settlement.Nonce,
		},
	})
}

// blockchain/integration/api/governance.go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"blockchain/consensus/governance"
	"blockchain/consensus/pos"
)

// generateID generates a unique ID based on timestamp
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// validatorExists checks if a validator exists in the pool
func validatorExists(addr string, pool *pos.ValidatorPool) bool {
	for _, v := range *pool {
		if v.Address == addr {
			return true
		}
	}
	return false
}

// Add fields to APIServer to hold references to governance components
type ExtendedAPIServer struct {
	*APIServer
	ValidatorPool    *pos.ValidatorPool
	GovernanceManager *governance.GovernanceManager
}

func (s *ExtendedAPIServer) handleNewProposal(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Title       string                 `json:"title"`
		Description string                 `json:"description"`
		Type        string                 `json:"type"`
		Parameters  map[string]interface{} `json:"parameters"`
	}
	
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Получаем адрес автора из контекста или заголовка
	author := r.Header.Get("X-Validator-Address")
	if author == "" {
		http.Error(w, "Validator address not provided", http.StatusBadRequest)
		return
	}
	
	// Преобразуем тип предложения
	var proposalType governance.ProposalType
	switch req.Type {
	case "parameter_change":
		proposalType = governance.ParameterChange
	case "protocol_upgrade":
		proposalType = governance.ProtocolUpgrade
	case "funding_request":
		proposalType = governance.FundingRequest
	default:
		http.Error(w, "Invalid proposal type", http.StatusBadRequest)
		return
	}
	
	// Создаем новое предложение
	proposal := governance.NewProposal(
		"gov-"+generateID(),
		req.Title,
		req.Description,
		author,
		proposalType,
		0.67,
		s.ValidatorPool,
	)
	
	// Добавляем параметры
	proposal.Parameters = req.Parameters
	
	// Добавляем предложение в говернанс
	s.GovernanceManager.SubmitProposal(proposal)
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"proposal_id": proposal.ID,
	})
}

func (s *ExtendedAPIServer) handleVote(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		ProposalID string `json:"proposal_id"`
		Choice     string `json:"choice"`
	}
	
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Получаем адрес валидатора
	validatorAddr := r.Header.Get("X-Validator-Address")
	if validatorAddr == "" {
		http.Error(w, "Validator address not provided", http.StatusBadRequest)
		return
	}
	
	// Проверяем, что валидатор существует
	if !validatorExists(validatorAddr, s.ValidatorPool) {
		http.Error(w, "Invalid validator address", http.StatusBadRequest)
		return
	}
	
	// Голосуем
	s.GovernanceManager.VoteOnProposal(req.ProposalID, validatorAddr, req.Choice)
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func (s *ExtendedAPIServer) handleProposalDetails(w http.ResponseWriter, r *http.Request) {
	proposalID := r.URL.Query().Get("id")
	if proposalID == "" {
		http.Error(w, "Proposal ID not provided", http.StatusBadRequest)
		return
	}
	
	proposal := s.GovernanceManager.Proposals[proposalID]
	if proposal == nil {
		http.Error(w, "Proposal not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(proposal)
}

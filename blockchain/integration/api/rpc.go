package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type JSONRPCRequest struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type JSONRPCResponse struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

func (s *APIServer) handleRPC(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req JSONRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var result interface{}
	var err error

	switch req.Method {
	case "getBlockByNumber":
		block := s.Chain.GetBlockByNumber(req.Params[0])
		result = block
	default:
		err = fmt.Errorf("method not found")
	}

	if err != nil {
		json.NewEncoder(w).Encode(JSONRPCResponse{
			ID:    req.ID,
			Error: err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(JSONRPCResponse{
		ID:     req.ID,
		Result: result,
	})
}

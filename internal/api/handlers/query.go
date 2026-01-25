package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"meueu/internal/core/domain"
	"meueu/internal/storage/postgres"
)

type QueryHandler struct {
	repo *postgres.NodeRepository
}

func NewQueryHandler(repo *postgres.NodeRepository) *QueryHandler {
	return &QueryHandler{repo: repo}
}

func (h *QueryHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}

	nodes, err := h.repo.Query(r.Context(), req.Filters, req.Limit, req.Offset)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		fmt.Printf("Query Error: %v\n", err)
		return
	}

	// Always return empty array instead of null
	if nodes == nil {
		nodes = []*domain.Node{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

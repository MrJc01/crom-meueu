package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"meueu/internal/storage/postgres"
)

type PeerHandler struct {
	repo *postgres.PeerRepository
}

func NewPeerHandler(repo *postgres.PeerRepository) *PeerHandler {
	return &PeerHandler{repo: repo}
}

func (h *PeerHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Dev mode: no DB connection
	if h.repo == nil || !h.repo.IsAvailable() {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	// Active peers in the last 24 hours
	cutoff := time.Now().Add(-24 * time.Hour)
	peers, err := h.repo.ListActive(r.Context(), cutoff, 50)
	if err != nil {
		http.Error(w, "Failed to fetch peers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peers)
}

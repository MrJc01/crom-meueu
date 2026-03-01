package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"meueu/internal/core/domain"
	"meueu/internal/storage/postgres"
)

// SyncHandler serves recent posts to other nodes for content synchronization.
type SyncHandler struct {
	repo *postgres.NodeRepository
}

func NewSyncHandler(repo *postgres.NodeRepository) *SyncHandler {
	return &SyncHandler{repo: repo}
}

// Handle serves GET /v1/sync?since=<unix_timestamp>&limit=<n>
// Returns up to `limit` posts with claimed_at > since, ordered by claimed_at ASC.
func (h *SyncHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse 'since' parameter (Unix timestamp)
	sinceStr := r.URL.Query().Get("since")
	var since time.Time
	if sinceStr != "" {
		sinceUnix, err := strconv.ParseInt(sinceStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid 'since' parameter: must be Unix timestamp", http.StatusBadRequest)
			return
		}
		since = time.Unix(sinceUnix, 0)
	} else {
		// Default: last 24 hours
		since = time.Now().Add(-24 * time.Hour)
	}

	// Parse 'limit' parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	nodes, err := h.repo.ListRecent(r.Context(), since, limit)
	if err != nil {
		http.Error(w, "Failed to fetch recent nodes", http.StatusInternalServerError)
		return
	}

	if nodes == nil {
		nodes = []*domain.Node{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

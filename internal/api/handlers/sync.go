package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"meueu/internal/core/domain"
	"meueu/internal/core/security"
	"meueu/internal/storage/postgres"
)

// SyncHandler serves recent posts to other nodes for content synchronization.
type SyncHandler struct {
	repo      *postgres.NodeRepository
	adminRepo *postgres.AdminRepository
}

func NewSyncHandler(repo *postgres.NodeRepository, adminRepo *postgres.AdminRepository) *SyncHandler {
	return &SyncHandler{
		repo:      repo,
		adminRepo: adminRepo,
	}
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

// HandleImport serves POST /v1/import
// Accepts a list of historic nodes (usually a backup or sync batch) and inserts them transactionally.
func (h *SyncHandler) HandleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10485760) // 10MB limit for bulk backups
	defer r.Body.Close()

	var nodes []*domain.Node
	if err := json.NewDecoder(r.Body).Decode(&nodes); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if len(nodes) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	expectedNetworkID := os.Getenv("NETWORK_ID")
	if expectedNetworkID == "" {
		expectedNetworkID = "meueu-mainnet-v1"
	}

	// Validation Loop
	validNodes := make([]*domain.Node, 0, len(nodes))
	for _, node := range nodes {
		if node.NetworkID != expectedNetworkID {
			continue // ignore wrong network silently in batch
		}

		payloadBytes, err := node.Payload.MarshalJSON()
		if err != nil { continue }

		// Needs security import
		// verify signature math
		valid, err := security.VerifySignature(node.AuthorPubkey, node.Signature, node.Kind, node.ClaimedAt.Unix(), node.Nonce, node.NetworkID, payloadBytes)
		if err != nil || !valid {
			continue
		}

		// Calculate ID 
		sigHash := sha256.Sum256([]byte(node.Signature))
		sigHashHex := fmt.Sprintf("%x", sigHash)

		// Check Banlist
		if h.adminRepo != nil {
			isBanned, err := h.adminRepo.IsBannedHash(r.Context(), sigHashHex)
			if err == nil && isBanned {
				continue
			}
		}

		var idBytes [16]byte
		copy(idBytes[:], sigHash[:16])
		idBytes[6] = (idBytes[6] & 0x0f) | 0x80
		idBytes[8] = (idBytes[8] & 0x3f) | 0x80
		
		node.ID = uuid.UUID(idBytes)
		node.OriginServer = r.Host
		
		validNodes = append(validNodes, node)
	}

	if len(validNodes) > 0 {
		if err := h.repo.CreateBatch(r.Context(), validNodes); err != nil {
			http.Error(w, "Transaction failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"imported": len(validNodes),
		"ignored": len(nodes) - len(validNodes),
	})
}

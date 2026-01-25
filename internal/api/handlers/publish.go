package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"meueu/internal/core/domain"
	"meueu/internal/core/security"
	"meueu/internal/storage/postgres"

	"github.com/google/uuid"
)

type PublishHandler struct {
	repo *postgres.NodeRepository
}

func NewPublishHandler(repo *postgres.NodeRepository) *PublishHandler {
	return &PublishHandler{repo: repo}
}

func (h *PublishHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.Node
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// 1. Basic Validation
	if req.AuthorPubkey == "" || req.Kind == "" || req.Signature == "" {
		http.Error(w, "Missing required fields (author_pubkey, kind, signature)", http.StatusBadRequest)
		return
	}

	// 2. Security: Construct Canonical String for Verification
	// Format: version:1|author:<pubkey>|kind:<kind>|timestamp:<claimed_at_unix>|payload:<payload_json_string>
	payloadBytes, _ := req.Payload.MarshalJSON() // Ensure we get the raw JSON string
	canonicalMsg := fmt.Sprintf("version:1|author:%s|kind:%s|timestamp:%d|payload:%s",
		req.AuthorPubkey,
		req.Kind,
		req.ClaimedAt.Unix(),
		string(payloadBytes),
	)

	// 3. Verify Signature
	valid, err := security.VerifySignature(req.AuthorPubkey, req.Signature, []byte(canonicalMsg))
	if err != nil {
		http.Error(w, fmt.Sprintf("Signature verification error: %v", err), http.StatusBadRequest)
		return
	}
	if !valid {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// 4. Generate ID (Server Authority)
	// In a fully decentralized p2p system, the ID might be the deterministic hash of the content.
	// For this protocol version, we assign a UUID v7 (time-based) or v4.
	req.ID = uuid.New()

	// 5. Persist
	if err := h.repo.Create(r.Context(), &req); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		// TODO: Log real error to stdout
		fmt.Printf("DB Error: %v\n", err)
		return
	}

	// 6. Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id":     req.ID.String(),
		"status": "published",
	})
}

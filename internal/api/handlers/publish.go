package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

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

	// 1. Security: Anti-DoS (Memory Exhaustion)
	// Enforce strict limit on request body size (1MB).
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		if errors.As(err, new(*http.MaxBytesError)) {
			http.Error(w, "Request body too large (limit 1MB)", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var req domain.Node
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// 2. Basic Validation
	if req.AuthorPubkey == "" || req.Kind == "" || req.Signature == "" {
		http.Error(w, "Missing required fields (author_pubkey, kind, signature)", http.StatusBadRequest)
		return
	}
	if req.NetworkID == "" || req.Nonce == "" {
		http.Error(w, "Missing required fields (network_id, nonce)", http.StatusBadRequest)
		return
	}

	// 3. Security: Time Travel & Import Logic
	now := time.Now()
	drift := req.ClaimedAt.Sub(now)

	// Rule: Future dates > 5m are strictly forbidden (Time Travelers/Clock Skew Attacks).
	if drift > 5*time.Minute {
		http.Error(w, fmt.Sprintf("Invalid timestamp: claimed_at is in the future (%v). Clock skew > 5m not allowed.", req.ClaimedAt), http.StatusBadRequest)
		return
	}
	// Past dates (Import Mode) are implicitly allowed.

	// 4. Verify Signature (V2 Strict TLV)
	payloadBytes, err := req.Payload.MarshalJSON()
	if err != nil {
		http.Error(w, "Invalid payload format", http.StatusBadRequest)
		return
	}

	valid, err := security.VerifySignature(req.AuthorPubkey, req.Signature, req.Kind, req.ClaimedAt.Unix(), req.Nonce, req.NetworkID, payloadBytes)
	if err != nil {
		// Return 400 for bad signature format/checks
		http.Error(w, fmt.Sprintf("Signature verification error: %v", err), http.StatusBadRequest)
		return
	}
	if !valid {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// 5. Generate ID (Content-Addressable / Deterministic)
	// ID = SHA256(Signature)
	sigHash := sha256.Sum256([]byte(req.Signature))
	var idBytes [16]byte
	copy(idBytes[:], sigHash[:16])

	// UUID v8-like (Custom)
	idBytes[6] = (idBytes[6] & 0x0f) | 0x80 // Version 8
	idBytes[8] = (idBytes[8] & 0x3f) | 0x80 // Variant 10
	req.ID = uuid.UUID(idBytes)

	// 6. Check Idempotency
	existingNode, err := h.repo.GetByID(r.Context(), req.ID)
	if err == nil && existingNode != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"id":     req.ID.String(),
			"status": "already_exists",
		})
		return
	}

	// 7. Persist
	// IMPORTANT: VerifiedAt is ALWAYS set to NOW using the server's clock.
	// This proves WHEN it was seen by the network, regardless of claimed_at.
	req.VerifiedAt = now

	if err := h.repo.Create(r.Context(), &req); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"id":     req.ID.String(),
				"status": "already_exists",
			})
			return
		}
		fmt.Printf("DB Error: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id":     req.ID.String(),
		"status": "published",
	})
}

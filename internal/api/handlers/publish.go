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
	// This prevents OOM attacks by reading unlimited streams.
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
	// Rule 1: Future dates > 5m are strictly forbidden (Time Travelers/Clock Skew Attacks).
	// Rule 2: Past dates are ALLOWED as "Historic Imports".
	// Rule 3: `verified_at` is ALWAYS `now` to track ingestion time.
	now := time.Now()
	drift := req.ClaimedAt.Sub(now)

	if drift > 5*time.Minute {
		// Future: Reject
		http.Error(w, fmt.Sprintf("Invalid timestamp: claimed_at is in the future (%v). Clock skew > 5m not allowed.", req.ClaimedAt), http.StatusBadRequest)
		return
	}

	// If message is older than 5 minutes, we consider it an "Import".
	// No error, just standard processing.

	// 4. Verify Signature (V2 Strict + Sanitized)
	// We pass raw payload to let VerifySignature handle canonicalization.
	// Re-marshalling payload ensures we verify what we parsed, although using original bytes slice for payload logic
	// would may be safer if JSON field ordering matters.
	// However, `req.Payload` is `json.RawMessage` or similar, usually a map/struct.
	// We trust `req.Payload.MarshalJSON()` produces the canonical payload bytes expected by the signer.
	// Note: Ideally, the client sends the payload string exactly as signed.
	payloadBytes, err := req.Payload.MarshalJSON()
	if err != nil {
		http.Error(w, "Invalid payload format", http.StatusBadRequest)
		return
	}

	valid, err := security.VerifySignature(req.AuthorPubkey, req.Signature, req.Kind, req.ClaimedAt.Unix(), req.Nonce, req.NetworkID, payloadBytes)
	if err != nil {
		// Log specific signature error for debugging, return generic bad request
		// fmt.Printf("Sig Debug: %v\n", err)
		http.Error(w, fmt.Sprintf("Signature verification error: %v", err), http.StatusBadRequest)
		return
	}
	if !valid {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// 5. Generate ID (Content-Addressable / Deterministic)
	// ID = SHA256(Signature). This guarantees that:
	// - Uniqueness changes if Sig changes (which changes if Content/Nonce/Time changes).
	// - Re-submitting the EXACT SAME signed message results in the SAME ID. (Idempotency Key)
	sigHash := sha256.Sum256([]byte(req.Signature))
	var idBytes [16]byte
	copy(idBytes[:], sigHash[:16])

	// UUID v8-like (Custom)
	idBytes[6] = (idBytes[6] & 0x0f) | 0x80 // Version 8
	idBytes[8] = (idBytes[8] & 0x3f) | 0x80 // Variant 10
	req.ID = uuid.UUID(idBytes)

	// 6. Security: Replay Attack Defense (Idempotency)
	// Check if ID exists.
	existingNode, err := h.repo.GetByID(r.Context(), req.ID) // Assuming GetByID exists or using a quick check
	if err == nil && existingNode != nil {
		// ID exists. This is a Replay or Retry.
		// Return 200 OK (Idempotent) to client, but do not process/store again.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"id":     req.ID.String(),
			"status": "already_exists",
		})
		return
	}

	// 7. Persist
	req.VerifiedAt = now

	if err := h.repo.Create(r.Context(), &req); err != nil {
		// Handle race condition if two requests came in parallel and both passed the 'GetByID' check
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

	// 8. Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id":     req.ID.String(),
		"status": "published",
	})
}

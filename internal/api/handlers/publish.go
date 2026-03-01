package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"meueu/internal/core/domain"
	"meueu/internal/core/security"
	"meueu/internal/storage/postgres"

	"github.com/google/uuid"
)

type PublishHandler struct {
	repo      *postgres.NodeRepository
	nonceRepo *postgres.NonceRepository
}

func NewPublishHandler(repo *postgres.NodeRepository, nonceRepo *postgres.NonceRepository) *PublishHandler {
	return &PublishHandler{
		repo:      repo,
		nonceRepo: nonceRepo,
	}
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

	// 2.5 Security: Network ID Verification
	expectedNetworkID := os.Getenv("NETWORK_ID")
	if expectedNetworkID == "" {
		expectedNetworkID = "meueu-mainnet-v1" // Fallback
	}
	if req.NetworkID != expectedNetworkID {
		http.Error(w, fmt.Sprintf("Invalid Network ID. Expected: %s", expectedNetworkID), http.StatusBadRequest)
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
	// Check if this server accepts external sync/imports.
	if drift < -5*time.Minute {
		syncExternal := os.Getenv("SYNC_EXTERNAL_POSTS")
		if syncExternal == "false" {
			http.Error(w, "This server does not accept historic imports (SYNC_EXTERNAL_POSTS=false).", http.StatusForbidden)
			return
		}
	}

	// 4. Security: Replay Attack Protection (Nonce Check)
	// Prevent unauthorized re-broadcast of the same valid message/nonce.
	if h.isReplayDetected(r.Context(), req.AuthorPubkey, req.Nonce) {
		http.Error(w, "Replay detected: Nonce already used.", http.StatusBadRequest)
		return
	}

	// 5. Verify Signature (V2 Strict + Sanitized)
	// We pass raw payload to let VerifySignature handle canonicalization.
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

	// 6. Generate ID (Content-Addressable / Deterministic)
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

	// 7. Check Idempotency (Existing ID)
	existingNode, err := h.repo.GetByID(r.Context(), req.ID)
	if err == nil && existingNode != nil {
		// ID exists. This is a Resubmission (Idempotent).
		// We return 200 OK to indicate "Processed".
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"id":     req.ID.String(),
			"status": "already_exists",
		})
		return
	}

	// 8. Persist
	// IMPORTANT: VerifiedAt is ALWAYS set to NOW using the server's clock.
	req.VerifiedAt = now

	if err := h.repo.Create(r.Context(), &req); err != nil {
		// Handle race condition
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

	// 9. Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id":     req.ID.String(),
		"status": "published",
	})
}

// isReplayDetected checks if the nonce has been used before.
// We use Postgres to atomically test-and-set the nonce.
func (h *PublishHandler) isReplayDetected(ctx context.Context, authorPubkey, nonce string) bool {
	if h.nonceRepo == nil {
		return false // Failsafe if not wired
	}

	success := h.nonceRepo.MarkUsed(ctx, nonce, authorPubkey)
	// If NOT successful, it means the nonce already exists -> Replay Detected!
	return !success
}

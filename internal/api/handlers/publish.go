package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
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

	// 2. Security: Drift Window Check (±5 minutes)
	// Prevents temporal paradoxes and some replay scenarios
	now := time.Now()
	drift := req.ClaimedAt.Sub(now)
	if drift > 5*time.Minute || drift < -5*time.Minute {
		http.Error(w, fmt.Sprintf("Time drift detected: claimed_at is %v, server time is %v. Allowed window is ±5m.", req.ClaimedAt, now), http.StatusBadRequest)
		return
	}

	// 3. Verify Signature (V2 Strict)
	// We pass raw payload to let VerifySignature handle canonicalization
	payloadBytes, _ := req.Payload.MarshalJSON()
	valid, err := security.VerifySignature(req.AuthorPubkey, req.Signature, req.Kind, req.ClaimedAt.Unix(), payloadBytes)
	if err != nil {
		http.Error(w, fmt.Sprintf("Signature verification error: %v", err), http.StatusBadRequest)
		return
	}
	if !valid {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// 4. Generate ID (Content-Addressable / Deterministic)
	// We use the SHA256 of the signature to ensure the ID is strictly bound to the authenticated content.
	// Since ID is a UUID (16 bytes), we take the first 16 bytes of the SHA256 hash.
	sigHash := sha256.Sum256([]byte(req.Signature))
	var idBytes [16]byte
	copy(idBytes[:], sigHash[:16])

	// Set version and variant bits to make it a valid UUIDv8-like (custom) or just raw bytes
	// UUID v4 is random, v5 is SHA1. Here we just use the bytes as identifier.
	// To comply with RFC 4122 slightly better, we could set version 0b1000 (8) but standard lib strictness varies.
	// For now, raw bytes ensure 128-bit security subset of SHA256.
	// But let's set the Version to 8 (custom) and Variant to RFC 4122 (10xx) to avoid confusion if parsers check it.
	idBytes[6] = (idBytes[6] & 0x0f) | 0x80 // Version 8
	idBytes[8] = (idBytes[8] & 0x3f) | 0x80 // Variant 10

	req.ID = uuid.UUID(idBytes)

	// 5. Persist
	// We verify the verified_at time to now
	req.VerifiedAt = now

	if err := h.repo.Create(r.Context(), &req); err != nil {
		// Log real error to stdout but hide from user
		fmt.Printf("DB Error: %v\n", err)

		// Check for duplicate key (idempotency)
		// If ID already exists, we should return 200 OK (idempotent) or 409 Conflict
		// For P2P, seeing the same message again is normal.
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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

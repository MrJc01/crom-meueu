package handlers

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"meueu/internal/core/identity"
	"meueu/internal/storage/postgres"
)

type PeerHandler struct {
	repo     *postgres.PeerRepository
	serverID identity.ServerIdentity
}

// SignedPeerList is the payload wrapper asserting authenticity
type SignedPeerList struct {
	Peers     interface{} `json:"peers"`
	Timestamp int64       `json:"timestamp"`
	Signature string      `json:"signature"`
	PubKey    string      `json:"pubkey"`
}

func NewPeerHandler(repo *postgres.PeerRepository, serverID identity.ServerIdentity) *PeerHandler {
	return &PeerHandler{repo: repo, serverID: serverID}
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

	// Sign the payload
	timestamp := time.Now().Unix()
	
	// Prepare the raw data to be signed: "peers_json|timestamp"
	peersJSON, _ := json.Marshal(peers)
	dataToSign := fmt.Sprintf("%s|%d", string(peersJSON), timestamp)
	
	// Hash data
	hash := sha256.Sum256([]byte(dataToSign))
	
	// Sign hash
	signature := ed25519.Sign(h.serverID.PrivateKey, hash[:])
	
	payload := SignedPeerList{
		Peers:     peers,
		Timestamp: timestamp,
		Signature: hex.EncodeToString(signature),
		PubKey:    h.serverID.PubKeyHex,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

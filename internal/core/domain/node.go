package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Node represents the atomic unit of content in the Crom Social Protocol.
// It maps directly to the 'nodes' table in the database.
type Node struct {
	ID           uuid.UUID       `json:"id"`
	ParentID     *uuid.UUID      `json:"parent_id,omitempty"` // Pointer to allow null
	AuthorPubkey string          `json:"author_pubkey"`
	Kind         string          `json:"kind"`
	Payload      json.RawMessage `json:"payload"`
	Tags         json.RawMessage `json:"tags,omitempty"`
	Signature    string          `json:"signature"`
	NetworkID    string          `json:"network_id"` // Anti-replay: domain separation
	Nonce        string          `json:"nonce"`      // Anti-replay: uniqueness
	ClaimedAt    time.Time       `json:"claimed_at"`
	VerifiedAt   time.Time       `json:"verified_at"`
	OriginServer string          `json:"origin_server,omitempty"`
}

// NewNode creates a basic Node structure.
// Note: ID generation and Signature must be handled by the caller/client appropriately.
func NewNode(authorPubkey, kind string, payload json.RawMessage) *Node {
	return &Node{
		ID:           uuid.New(), // Will be overwritten by hash(signature) in newer protocol versions
		AuthorPubkey: authorPubkey,
		Kind:         kind,
		Payload:      payload,
		NetworkID:    "meueu-mainnet-v1",  // Default to mainnet
		Nonce:        uuid.New().String(), // Random nonce
		ClaimedAt:    time.Now(),
	}
}

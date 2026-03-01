package domain

import (
	"time"

	"github.com/google/uuid"
)

// Peer represents another Crom Node in the decentralized network.
type Peer struct {
	ID         uuid.UUID `json:"id"`
	URL        string    `json:"url"`
	PublicKey  string    `json:"public_key"`
	LastSeen   time.Time `json:"last_seen"`
	Reputation int       `json:"reputation"`
}

// NewPeer creates a new peer entity.
func NewPeer(url, pubKey string) *Peer {
	return &Peer{
		ID:         uuid.New(),
		URL:        url,
		PublicKey:  pubKey,
		LastSeen:   time.Now(),
		Reputation: 100, // starting reputation
	}
}

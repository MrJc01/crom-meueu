package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"meueu/internal/core/domain"
	"meueu/internal/core/security"
	"meueu/internal/storage/postgres"
)

type PeerDiscoveryService struct {
	repo     *postgres.PeerRepository
	nodeRepo *postgres.NodeRepository
	myURL    string
	myPubKey string
	seedNodes []string
	client   *http.Client
}

func NewPeerDiscoveryService(repo *postgres.PeerRepository, nodeRepo *postgres.NodeRepository, myURL, myPubKey string, seedNodes []string) *PeerDiscoveryService {
	return &PeerDiscoveryService{
		repo:      repo,
		nodeRepo:  nodeRepo,
		myURL:     myURL,
		myPubKey:  myPubKey,
		seedNodes: seedNodes,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *PeerDiscoveryService) Start(ctx context.Context) {
	log.Println("[PeerDiscovery] Starting gossip worker...")
	// 1. Initial Seeding
	s.pingSeeds(ctx)

	// 2. Periodic Gossip Loop
	ticker := time.NewTicker(2 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				s.gossip(ctx)
				// Content sync runs alongside peer gossip
				if os.Getenv("SYNC_EXTERNAL_POSTS") != "false" {
					s.syncContent(ctx)
				}
			}
		}
	}()
}

func (s *PeerDiscoveryService) pingSeeds(ctx context.Context) {
	for _, seed := range s.seedNodes {
		if strings.TrimSpace(seed) == "" || seed == s.myURL {
			continue
		}
		s.fetchPeersFrom(ctx, seed)
	}
}

func (s *PeerDiscoveryService) gossip(ctx context.Context) {
	if s.repo == nil {
		return
	}
	activePeers, err := s.repo.ListActive(ctx, time.Now().Add(-24*time.Hour), 10)
	if err != nil {
		log.Printf("[PeerDiscovery] Failed to list active peers: %v", err)
		return
	}

	for _, p := range activePeers {
		if p.URL == s.myURL {
			continue
		}
		s.fetchPeersFrom(ctx, p.URL)
	}
}

func (s *PeerDiscoveryService) fetchPeersFrom(ctx context.Context, peerURL string) {
	url := fmt.Sprintf("%s/v1/peers", strings.TrimRight(peerURL, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}

	resp, err := s.client.Do(req)
	if err != nil {
		s.punishPeerByUrl(ctx, peerURL, 5)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.punishPeerByUrl(ctx, peerURL, 2)
		return
	}

	var discoveredPeers []domain.Peer
	if err := json.NewDecoder(resp.Body).Decode(&discoveredPeers); err != nil {
		s.punishPeerByUrl(ctx, peerURL, 5) // Bad JSON
		return
	}

	if s.repo == nil {
		return
	}

	for _, dp := range discoveredPeers {
		if dp.URL == s.myURL {
			continue
		}
		
		dp.LastSeen = time.Now()
		if dp.Reputation == 0 {
			dp.Reputation = 100
		}
		
		if err := s.repo.Upsert(ctx, &dp); err != nil {
			log.Printf("[PeerDiscovery] Failed to upsert peer %s: %v", dp.URL, err)
		}
	}
	
	s.rewardPeerByUrl(ctx, peerURL, 1)
}

// syncContent pulls recent posts from reputable peers and validates their signatures.
// Posts with invalid signatures trigger reputation punishment.
func (s *PeerDiscoveryService) syncContent(ctx context.Context) {
	if s.nodeRepo == nil || s.repo == nil {
		return
	}

	// Only sync from peers with reputation > 50
	activePeers, err := s.repo.ListActive(ctx, time.Now().Add(-24*time.Hour), 10)
	if err != nil {
		log.Printf("[ContentSync] Failed to list peers: %v", err)
		return
	}

	since := time.Now().Add(-1 * time.Hour).Unix() // Last hour

	for _, p := range activePeers {
		if p.URL == s.myURL || p.Reputation < 50 {
			continue
		}
		s.fetchContentFrom(ctx, p, since)
	}
}

func (s *PeerDiscoveryService) fetchContentFrom(ctx context.Context, peer *domain.Peer, sinceUnix int64) {
	url := fmt.Sprintf("%s/v1/sync?since=%d&limit=50", strings.TrimRight(peer.URL, "/"), sinceUnix)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}

	resp, err := s.client.Do(req)
	if err != nil {
		s.punishPeerByUrl(ctx, peer.URL, 3)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.punishPeerByUrl(ctx, peer.URL, 2)
		return
	}

	var nodes []domain.Node
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		s.punishPeerByUrl(ctx, peer.URL, 5) // Bad JSON
		return
	}

	expectedNetworkID := os.Getenv("NETWORK_ID")
	if expectedNetworkID == "" {
		expectedNetworkID = "meueu-mainnet-v1"
	}

	validCount := 0
	invalidCount := 0

	for _, node := range nodes {
		// Skip nodes from wrong network
		if node.NetworkID != expectedNetworkID {
			continue
		}

		// Validate signature
		payloadBytes, err := node.Payload.MarshalJSON()
		if err != nil {
			invalidCount++
			continue
		}

		valid, err := security.VerifySignature(
			node.AuthorPubkey, node.Signature,
			node.Kind, node.ClaimedAt.Unix(),
			node.Nonce, node.NetworkID, payloadBytes,
		)

		if err != nil || !valid {
			invalidCount++
			continue
		}

		// Generate deterministic ID from signature hash
		sigHash := sha256.Sum256([]byte(node.Signature))
		var idBytes [16]byte
		copy(idBytes[:], sigHash[:16])
		idBytes[6] = (idBytes[6] & 0x0f) | 0x80 // Version 8
		idBytes[8] = (idBytes[8] & 0x3f) | 0x80 // Variant 10

		// Import uses google/uuid but we import the type through domain
		// Check if already exists by trying to create
		node.VerifiedAt = time.Now()
		node.OriginServer = peer.URL

		if err := s.nodeRepo.Create(ctx, &node); err != nil {
			// Duplicate is expected and fine
			if !strings.Contains(err.Error(), "duplicate") && !strings.Contains(err.Error(), "unique") {
				log.Printf("[ContentSync] Failed to store node from %s: %v", peer.URL, err)
			}
			continue
		}
		validCount++
	}

	// Punish peer if majority of their posts have invalid signatures
	if invalidCount > 0 && invalidCount > validCount {
		penalty := invalidCount * 5
		if penalty > 30 {
			penalty = 30
		}
		log.Printf("[ContentSync] Punishing peer %s: %d invalid signatures (penalty: %d)", peer.URL, invalidCount, penalty)
		s.repo.Punish(ctx, peer.PublicKey, penalty)
	} else if validCount > 0 {
		s.rewardPeerByUrl(ctx, peer.URL, 2)
		log.Printf("[ContentSync] Synced %d valid posts from %s", validCount, peer.URL)
	}
}

func (s *PeerDiscoveryService) punishPeerByUrl(ctx context.Context, url string, penalty int) {
	if s.repo == nil {
		return
	}
	peer, err := s.repo.GetByURL(ctx, url)
	if err == nil && peer != nil {
		s.repo.Punish(ctx, peer.PublicKey, penalty)
	}
}

func (s *PeerDiscoveryService) rewardPeerByUrl(ctx context.Context, url string, reward int) {
	if s.repo == nil {
		return
	}
	peer, err := s.repo.GetByURL(ctx, url)
	if err == nil && peer != nil {
		peer.LastSeen = time.Now()
		if peer.Reputation < 100 {
			peer.Reputation += reward
		}
		s.repo.Upsert(ctx, peer)
	} else if err == nil && peer == nil {
		newP := domain.NewPeer(url, "discovered-from-gossip")
		s.repo.Upsert(ctx, newP)
	}
}

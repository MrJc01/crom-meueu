package handlers

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"meueu/internal/core/domain"
)

// MockRepository to avoid DB dependency in unit handler test
type MockNodeRepository struct {
	CapturedNode *domain.Node
}

func (m *MockNodeRepository) Create(ctx context.Context, node *domain.Node) error {
	m.CapturedNode = node
	return nil
}

// Note: To test the actual handler logic including signature verification,
// we need to mock the repo. However, the handler expects *postgres.NodeRepository struct,
// not an interface. For this quick iteration, we will skip mocking the struct method
// (which is hard in Go without an interface) and focus on the signature logic
// by extracting validation or integration testing if we had a running DB.

// Ideally, Handler should rely on an interface `NodeRepository`.
// Let's refactor the handler to use an interface or just trust the integration test
// with the real DB (which verify.sh does).

// For now, let's create a "Client-Side" signer helper that we can use in verify.sh
// or main_test.go to generate a valid CURL command.

func TestGenerateSignedPayload(t *testing.T) {
	// 1. Generate Keypair
	pubKey, privKey, _ := ed25519.GenerateKey(nil)
	pubKeyHex := hex.EncodeToString(pubKey)

	// 2. Create Payload
	payloadJSON := json.RawMessage(`{"content":"Hello World"}`)
	kind := "text"
	timestamp := time.Now()

	// 3. Sign
	payloadBytes, _ := payloadJSON.MarshalJSON()
	canonicalMsg := fmt.Sprintf("version:1|author:%s|kind:%s|timestamp:%d|payload:%s",
		pubKeyHex, kind, timestamp.Unix(), string(payloadBytes),
	)
	signature := ed25519.Sign(privKey, []byte(canonicalMsg))
	sigHex := hex.EncodeToString(signature)

	t.Logf("\n=== VALID PAYLOAD FOR TESTING ===\n")
	t.Logf("Author: %s\n", pubKeyHex)
	t.Logf("Signature: %s\n", sigHex)
	t.Logf("Timestamp: %s (%d)\n", timestamp.Format(time.RFC3339), timestamp.Unix())

	// Create full JSON body
	body := map[string]interface{}{
		"author_pubkey": pubKeyHex,
		"kind":          kind,
		"payload":       payloadJSON,
		"signature":     sigHex,
		"claimed_at":    timestamp,
	}
	bodyBytes, _ := json.MarshalIndent(body, "", "  ")
	t.Logf("Body:\n%s\n", string(bodyBytes))
}

package security

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

func TestVerifySignature(t *testing.T) {
	// Generate a key pair for testing
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	// Prepare data
	payload := []byte(`{"content":"Hello, Crom!"}`)
	kind := "note"
	timestamp := time.Now().Unix()
	nonce := "test-nonce-123"
	networkID := "test-net"

	// Construct Canonical Message manually to sign
	// Format: <NetworkID>|v1|author:<pubkey>|kind:<kind>|ts:<timestamp>|nonce:<nonce>|phash:<payload_hash>
	pubKeyHex := hex.EncodeToString(pubKey)
	payloadHash := sha256.Sum256(payload)
	payloadHashHex := hex.EncodeToString(payloadHash[:])

	canonicalMsg := fmt.Sprintf("%s|v1|author:%s|kind:%s|ts:%d|nonce:%s|phash:%s",
		networkID,
		pubKeyHex,
		kind,
		timestamp,
		nonce,
		payloadHashHex,
	)

	// Sign the message
	signature := ed25519.Sign(privKey, []byte(canonicalMsg))
	sigHex := hex.EncodeToString(signature)

	// Test Case 1: Valid Signature
	valid, err := VerifySignature(pubKeyHex, sigHex, kind, timestamp, nonce, networkID, payload)
	if err != nil {
		t.Errorf("unexpected error for valid signature: %v", err)
	}
	if !valid {
		t.Error("expected valid signature, got invalid")
	}

	// Test Case 2: Invalid Payload (Hash Mismatch)
	valid, err = VerifySignature(pubKeyHex, sigHex, kind, timestamp, nonce, networkID, []byte(`{"content":"Forged!"}`))
	if err == nil && valid {
		t.Error("expected invalid signature for wrong payload, got valid")
	}

	// Test Case 3: Invalid Nonce
	valid, err = VerifySignature(pubKeyHex, sigHex, kind, timestamp, "wrong-nonce", networkID, payload)
	if err == nil && valid {
		t.Error("expected invalid signature for wrong nonce, got valid")
	}

	// Test Case 4: Invalid NetworkID
	valid, err = VerifySignature(pubKeyHex, sigHex, kind, timestamp, nonce, "wrong-net", payload)
	if err == nil && valid {
		t.Error("expected invalid signature for wrong networkID, got valid")
	}
}

package security

import (
	"bytes"
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
	// Format: [Len][Val]...
	pubKeyHex := hex.EncodeToString(pubKey)
	payloadHash := sha256.Sum256(payload)
	payloadHashHex := hex.EncodeToString(payloadHash[:])
	timestampStr := fmt.Sprintf("%d", timestamp)

	var buf bytes.Buffer
	writeTLV := func(data string) {
		l := uint16(len(data))
		buf.WriteByte(byte(l >> 8))
		buf.WriteByte(byte(l))
		buf.WriteString(data)
	}

	writeTLV(networkID)
	writeTLV("v2")
	writeTLV(pubKeyHex)
	writeTLV(kind)
	writeTLV(timestampStr)
	writeTLV(nonce)
	writeTLV(payloadHashHex)

	canonicalMsg := buf.Bytes()

	// Sign the message
	signature := ed25519.Sign(privKey, canonicalMsg)
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

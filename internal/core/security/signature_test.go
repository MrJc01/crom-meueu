package security

import (
	"crypto/ed25519"
	"encoding/hex"
	"testing"
)

func TestVerifySignature(t *testing.T) {
	// Generate a key pair for testing
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	// Prepare a message
	message := []byte("Hello, Crom!")

	// Sign the message
	signature := ed25519.Sign(privKey, message)

	// Encode to Hex strings (as expected by our utility)
	pubKeyHex := hex.EncodeToString(pubKey)
	sigHex := hex.EncodeToString(signature)

	// Test Case 1: Valid Signature
	valid, err := VerifySignature(pubKeyHex, sigHex, message)
	if err != nil {
		t.Errorf("unexpected error for valid signature: %v", err)
	}
	if !valid {
		t.Error("expected valid signature, got invalid")
	}

	// Test Case 2: Invalid Message
	valid, err = VerifySignature(pubKeyHex, sigHex, []byte("Wrong Message"))
	if err == nil && valid {
		t.Error("expected invalid signature for wrong message, got valid")
	}

	// Test Case 3: Invalid Public Key Hex
	_, err = VerifySignature("invalid-hex", sigHex, message)
	if err == nil {
		t.Error("expected error for invalid public key hex, got none")
	}
}

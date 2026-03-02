package security

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

func TestVerifySignature_Valid(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	pubHex := hex.EncodeToString(pub)

	networkID := "meueu-testnet-1"
	kind := "post"
	timestamp := time.Now().Unix()
	nonce := "random_nonce_123"
	payload := []byte(`{"text":"Hello World"}`)

	// Construct exactly as the verifier expects
	payloadHashHex := hex.EncodeToString(func(b []byte) []byte {
		h := cryptoHash(b)
		return h[:]
	}(payload))

	var buf bytes.Buffer
	writeTLV := func(data string) {
		l := uint16(len(data))
		binary.Write(&buf, binary.BigEndian, l)
		buf.WriteString(data)
	}

	writeTLV(networkID)
	writeTLV("v2")
	writeTLV(pubHex)
	writeTLV(kind)
	writeTLV(fmt.Sprintf("%d", timestamp))
	writeTLV(nonce)
	writeTLV(payloadHashHex)

	sig := ed25519.Sign(priv, buf.Bytes())
	sigHex := hex.EncodeToString(sig)

	// Test
	valid, err := VerifySignature(pubHex, sigHex, kind, timestamp, nonce, networkID, payload)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !valid {
		t.Errorf("Expected signature to be valid")
	}
}

func TestVerifySignature_TamperedPayload(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	pubHex := hex.EncodeToString(pub)

	networkID := "meueu-testnet-1"
	kind := "post"
	timestamp := time.Now().Unix()
	nonce := "random_nonce_123"
	payload := []byte(`{"text":"Original"}`)
	fakePayload := []byte(`{"text":"Hacked"}`)

	payloadHashHex := hex.EncodeToString(func(b []byte) []byte { h := cryptoHash(b); return h[:] }(payload))

	var buf bytes.Buffer
	writeTLV := func(data string) {
		l := uint16(len(data))
		binary.Write(&buf, binary.BigEndian, l)
		buf.WriteString(data)
	}

	writeTLV(networkID)
	writeTLV("v2")
	writeTLV(pubHex)
	writeTLV(kind)
	writeTLV(fmt.Sprintf("%d", timestamp))
	writeTLV(nonce)
	writeTLV(payloadHashHex)

	sig := ed25519.Sign(priv, buf.Bytes())
	sigHex := hex.EncodeToString(sig)

	// Verify against Fake Payload
	valid, err := VerifySignature(pubHex, sigHex, kind, timestamp, nonce, networkID, fakePayload)
	if err != nil {
		t.Fatalf("Expected no error string (just false boolean), got: %v", err)
	}
	if valid {
		t.Errorf("Expected signature to be INVALID for tampered payload")
	}
}

func TestVerifySignature_WrongNetwork(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	pubHex := hex.EncodeToString(pub)
	sig := ed25519.Sign(priv, []byte("fake"))
	sigHex := hex.EncodeToString(sig)

	// Try invalid chars
	_, err := VerifySignature(pubHex, sigHex, "post", 123, "nonce", "meueu|hacked", []byte(""))
	if err == nil {
		t.Errorf("Expected error for invalid network ID characters")
	}
}

func cryptoHash(b []byte) [32]byte {
	return sha256.Sum256(b)
}

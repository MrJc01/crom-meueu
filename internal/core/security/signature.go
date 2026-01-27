package security

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
)

// NetworkID identifies the specific network domain to prevent cross-network replay attacks.
const NetworkID = "meueu-mainnet-v1"

// Whitelist regex for safe inputs (Alphanumeric, hyphen, underscore).
var safeInputRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)

// VerifySignature checks if the provided signature is valid for the message and public key.
// It uses a strict Length-Prefix (TLV) serialization format to conduct the check:
// [2b len][val]... for: NetworkID, Version("v2"), AuthorPubkey, Kind, Timestamp(string), Nonce, PayloadHash
func VerifySignature(pubKeyHex, sigHex string, kind string, timestamp int64, nonce string, networkID string, payload []byte) (bool, error) {
	if pubKeyHex == "" || sigHex == "" {
		return false, errors.New("public key and signature cannot be empty")
	}

	// 1. Basic Sanitization
	if !safeInputRegex.MatchString(networkID) {
		return false, errors.New("invalid network_id characters")
	}
	if !safeInputRegex.MatchString(nonce) {
		return false, errors.New("invalid nonce characters")
	}
	if !safeInputRegex.MatchString(kind) {
		return false, errors.New("invalid kind characters")
	}

	if networkID == "" {
		return false, errors.New("network_id is required")
	}

	// Decode keys
	pubKey, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return false, fmt.Errorf("invalid public key hex: %w", err)
	}

	signature, err := hex.DecodeString(sigHex)
	if err != nil {
		return false, fmt.Errorf("invalid signature hex: %w", err)
	}

	if len(pubKey) != ed25519.PublicKeySize {
		return false, fmt.Errorf("invalid public key length: expected %d, got %d", ed25519.PublicKeySize, len(pubKey))
	}

	if len(signature) != ed25519.SignatureSize {
		return false, fmt.Errorf("invalid signature length: expected %d, got %d", ed25519.SignatureSize, len(signature))
	}

	// Construct Canonical Message using strict TLV (Big Endian Length Prefix)
	payloadHash := sha256.Sum256(payload)
	payloadHashHex := hex.EncodeToString(payloadHash[:])
	timestampStr := fmt.Sprintf("%d", timestamp)

	var buf bytes.Buffer

	// Helper to write TLV
	writeTLV := func(data string) {
		l := uint16(len(data))
		// Write Length (16-bit Big Endian)
		binary.Write(&buf, binary.BigEndian, l)
		// Write Value
		buf.WriteString(data)
	}

	// Strict Order: NetworkID -> Version -> Author -> Kind -> Timestamp -> Nonce -> PayloadHash
	writeTLV(networkID)
	writeTLV("v2")      // Enforce Protocol Version v2
	writeTLV(pubKeyHex) // Author PubKey
	writeTLV(kind)
	writeTLV(timestampStr)
	writeTLV(nonce)
	writeTLV(payloadHashHex)

	canonicalMsg := buf.Bytes()

	// Verify
	valid := ed25519.Verify(pubKey, canonicalMsg, signature)
	return valid, nil
}

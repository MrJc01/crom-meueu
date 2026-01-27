package security

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
)

// NetworkID identifies the specific network domain to prevent cross-network replay attacks.
const NetworkID = "meueu-mainnet-v1"

// Whitelist regex for safe inputs (Alphanumeric, hyphen, underscore).
// Prohibits characters like '|', ':', which are used as delimiters.
var safeInputRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)

// VerifySignature checks if the provided signature is valid for the message and public key.
// It uses a strict canonical serialization format:
// Key:Value|Key:Value...
// SHA256(NetworkID | version:1 | author:<pubkey> | kind:<kind> | timestamp:<ts> | nonce:<nonce> | payload_hash:<sha256(payload)>)
func VerifySignature(pubKeyHex, sigHex string, kind string, timestamp int64, nonce string, networkID string, payload []byte) (bool, error) {
	if pubKeyHex == "" || sigHex == "" {
		return false, errors.New("public key and signature cannot be empty")
	}

	// 1. Strict Input Sanitization (Injection Prevention)
	if !safeInputRegex.MatchString(networkID) {
		return false, errors.New("invalid network_id characters")
	}
	if !safeInputRegex.MatchString(nonce) {
		return false, errors.New("invalid nonce characters")
	}
	if !safeInputRegex.MatchString(kind) {
		return false, errors.New("invalid kind characters")
	}

	// Default NetworkID enforcement
	if networkID == "" {
		return false, errors.New("network_id is required")
	}

	// Decode the hex strings
	pubKey, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return false, fmt.Errorf("invalid public key hex: %w", err)
	}

	signature, err := hex.DecodeString(sigHex)
	if err != nil {
		return false, fmt.Errorf("invalid signature hex: %w", err)
	}

	// Validate key lengths
	if len(pubKey) != ed25519.PublicKeySize {
		return false, fmt.Errorf("invalid public key length: expected %d, got %d", ed25519.PublicKeySize, len(pubKey))
	}

	if len(signature) != ed25519.SignatureSize {
		return false, fmt.Errorf("invalid signature length: expected %d, got %d", ed25519.SignatureSize, len(signature))
	}

	// Construct Canonical Message
	// We hash the payload first to ensure size consistency.
	payloadHash := sha256.Sum256(payload)
	payloadHashHex := hex.EncodeToString(payloadHash[:])

	// Strict format: <NetworkID>|v1|author:<pubkey>|kind:<kind>|ts:<timestamp>|nonce:<nonce>|phash:<payload_hash>
	// Using bytes.Buffer for efficiency.
	// Input values are now guaranteed to NOT contain '|' due to regex check above.
	var buf bytes.Buffer
	buf.WriteString(networkID)
	buf.WriteString("|v1|author:")
	buf.WriteString(pubKeyHex)
	buf.WriteString("|kind:")
	buf.WriteString(kind)
	buf.WriteString("|ts:")
	buf.WriteString(fmt.Sprintf("%d", timestamp))
	buf.WriteString("|nonce:")
	buf.WriteString(nonce)
	buf.WriteString("|phash:")
	buf.WriteString(payloadHashHex)

	canonicalMsg := buf.Bytes()

	// Verify using standard crypto/ed25519
	valid := ed25519.Verify(pubKey, canonicalMsg, signature)
	return valid, nil
}

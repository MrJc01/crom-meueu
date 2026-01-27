package security

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// NetworkID identifies the specific network domain to prevent cross-network replay attacks.
const NetworkID = "meueu-mainnet-v1"

// VerifySignature checks if the provided signature is valid for the message and public key.
// It uses a strict canonical serialization format:
// SHA256(NetworkID | version:1 | author:<pubkey> | kind:<kind> | timestamp:<ts> | nonce:<nonce> | payload_hash:<sha256(payload)>)
func VerifySignature(pubKeyHex, sigHex string, kind string, timestamp int64, nonce string, networkID string, payload []byte) (bool, error) {
	if pubKeyHex == "" || sigHex == "" {
		return false, errors.New("public key and signature cannot be empty")
	}

	// Default NetworkID if empty (backward compatibility or strict enforcement?)
	// Strict: Fail if empty.
	if networkID == "" {
		return false, errors.New("network_id is required")
	}
	if nonce == "" {
		return false, errors.New("nonce is required")
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

	// Strict format: <NetworkID>|v1|<pubkey>|<kind>|<ts>|<nonce>|<phash>
	// Using bytes.Buffer for efficiency and avoiding format string injection risks
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

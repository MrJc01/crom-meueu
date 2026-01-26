package security

import (
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
// SHA256(NetworkID | version:1 | author:<pubkey> | kind:<kind> | timestamp:<ts> | payload_hash:<sha256(payload)>)
func VerifySignature(pubKeyHex, sigHex string, kind string, timestamp int64, payload []byte) (bool, error) {
	if pubKeyHex == "" || sigHex == "" {
		return false, errors.New("public key and signature cannot be empty")
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
	// We hash the payload first to ensure size consistency and avoid memory exhaustion for large payloads during concat
	payloadHash := sha256.Sum256(payload)
	payloadHashHex := hex.EncodeToString(payloadHash[:])

	// Strict format with delimiters that cannot be injected easily if we control the structure
	// Format: <NetworkID>|v1|author:<pubkey>|kind:<kind>|ts:<timestamp>|phash:<payload_hash>
	canonicalMsg := fmt.Sprintf("%s|v1|author:%s|kind:%s|ts:%d|phash:%s",
		NetworkID,
		pubKeyHex,
		kind,
		timestamp,
		payloadHashHex,
	)

	// Verify using standard crypto/ed25519
	valid := ed25519.Verify(pubKey, []byte(canonicalMsg), signature)
	return valid, nil
}

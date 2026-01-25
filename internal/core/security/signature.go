package security

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
)

// VerifySignature checks if the provided signature is valid for the message and public key.
// It expects the public key and signature to be hex-encoded strings.
func VerifySignature(pubKeyHex, sigHex string, message []byte) (bool, error) {
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

	// Verify using standard crypto/ed25519
	valid := ed25519.Verify(pubKey, message, signature)
	return valid, nil
}

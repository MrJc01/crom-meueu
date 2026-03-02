package identity

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
)

// ServerIdentity holds the cryptographic identity of the node
type ServerIdentity struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
	PubKeyHex  string
}

// GenerateServerKeypair creates a new Ed25519 keypair
func GenerateServerKeypair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate keypair: %w", err)
	}
	return pub, priv, nil
}

// LoadServerKeypair reads a previously saved Ed25519 private key from disk
func LoadServerKeypair(filepath string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	
	if len(data) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid private key size")
	}
	
	return ed25519.PrivateKey(data), nil
}

// SaveServerKeypair saves an Ed25519 private key to disk with strict 0600 permissions
func SaveServerKeypair(filepath string, key ed25519.PrivateKey) error {
	return os.WriteFile(filepath, key, 0600)
}

// InitServerIdentity loads or generates the server's identity
func InitServerIdentity(filepath string) (ServerIdentity, error) {
	var pub ed25519.PublicKey
	var priv ed25519.PrivateKey

	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		log.Println("Server key not found, generating a new identity...")
		pub, priv, err = GenerateServerKeypair()
		if err != nil {
			return ServerIdentity{}, err
		}
		
		err = SaveServerKeypair(filepath, priv)
		if err != nil {
			log.Fatalf("Fatal Error: 'Permission Denied' or unable to write %s. Ensure directory is writable.", filepath)
			return ServerIdentity{}, err
		}
		log.Println("New server identity generated and securely saved.")
	} else if err != nil {
		return ServerIdentity{}, fmt.Errorf("error accessing %s: %w", filepath, err)
	} else {
		log.Println("Loading existing server identity...")
		priv, err = LoadServerKeypair(filepath)
		if err != nil {
			log.Fatalf("Fatal Error: Failed to load server key: %v", err)
			return ServerIdentity{}, err
		}
		
		// Derive public key from private key
		// In Ed25519, the public key is the trailing 32 bytes of the 64-byte private key
		pub = priv.Public().(ed25519.PublicKey)
	}

	pubHex := hex.EncodeToString(pub)
	
	// Create short prefix for display
	prefixLen := 8
	if len(pubHex) < 8 {
		prefixLen = len(pubHex)
	}

	log.Printf("\033[32mBooting... PK: %s\033[0m\n", pubHex[:prefixLen])

	return ServerIdentity{
		PublicKey:  pub,
		PrivateKey: priv,
		PubKeyHex:  pubHex,
	}, nil
}

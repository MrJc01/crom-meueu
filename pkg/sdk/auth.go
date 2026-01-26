package sdk

import (
	"crypto/ed25519"
	"crypto/sha512"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/curve25519"
)

// CromAuth handles Identity, Signing and Encryption.
type CromAuth struct {
	edPrivKey ed25519.PrivateKey
	edPubKey  ed25519.PublicKey

	boxPrivKey [32]byte
	boxPubKey  [32]byte
}

func NewCromAuth() *CromAuth {
	return &CromAuth{}
}

// Login derives the private key from the seed phrase using Argon2id.
// WARNING: This is a breaking change from the SHA256 implementation.
func (a *CromAuth) Login(seedPhrase string) (string, error) {
	// Use Argon2id for key stretching.
	// Parameters: time=1, memory=64MB, threads=4, keyLen=32
	// Salt is static because we need deterministic key generation from the seed phrase alone.
	salt := []byte("meueu-protocol-salt-v1-hardening")
	derivedSeed := argon2.IDKey([]byte(seedPhrase), salt, 1, 64*1024, 4, 32)

	privKey := ed25519.NewKeyFromSeed(derivedSeed)
	a.edPrivKey = privKey
	a.edPubKey = privKey.Public().(ed25519.PublicKey)

	if err := a.deriveEncryptionKeys(); err != nil {
		return "", err
	}
	return hex.EncodeToString(a.edPubKey), nil
}

func (a *CromAuth) deriveEncryptionKeys() error {
	// Ed25519 Priv -> Curve25519 Priv
	// Method: SHA512 of seed, clamp first 32 bytes.
	seed := a.edPrivKey[:32]
	h := sha512.New()
	h.Write(seed)
	digest := h.Sum(nil)

	digest[0] &= 248
	digest[31] &= 127
	digest[31] |= 64

	copy(a.boxPrivKey[:], digest[:32])

	// Curve25519 Pub from Priv
	curve25519.ScalarBaseMult(&a.boxPubKey, &a.boxPrivKey)
	return nil
}

func (a *CromAuth) Sign(message string) string {
	sig := ed25519.Sign(a.edPrivKey, []byte(message))
	return hex.EncodeToString(sig)
}

func (a *CromAuth) EncryptDM(recipientEdPubHex, message string) (string, string, error) {
	recipBytes, err := hex.DecodeString(recipientEdPubHex)
	if err != nil {
		return "", "", err
	}
	if len(recipBytes) != 32 {
		return "", "", errors.New("invalid recipient key length")
	}

	var edPub [32]byte
	copy(edPub[:], recipBytes)

	// var curvePub [32]byte

	// TODO: Implement Ed25519 -> Curve25519 conversion properly.
	// Currently stubbed because external dependencies (agl/ed25519) are failing to fetch.
	return "", "", errors.New("encryption not supported: missing ed2curve implementation")

	/*
		if !extra25519.PublicKeyToCurve25519(&curvePub, &edPub) {
			return "", "", errors.New("failed to convert public key")
		}

		var nonce [24]byte
		if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
			return "", "", err
		}

		msgBytes := []byte(message)
		ciphertext := box.Seal(nil, msgBytes, &nonce, &curvePub, &a.boxPrivKey)

		return hex.EncodeToString(ciphertext), hex.EncodeToString(nonce[:]), nil
	*/
}

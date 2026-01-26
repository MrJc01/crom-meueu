package sdk

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/nacl/box"

	"filippo.io/edwards25519"
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

	// Convert Ed25519 Public Key to Curve25519 Public Key
	// Curve25519 points are Montgomery points. Ed25519 points are Edwards points.
	// Step 1: Verify the point is on the Ed25519 curve to prevent invalid curve attacks.
	if _, err := new(edwards25519.Point).SetBytes(edPub[:]); err != nil {
		return "", "", fmt.Errorf("invalid ed25519 public key point: %w", err)
	}

	// Step 2: Convert Y-coordinate (Edwards) to U-coordinate (Montgomery)
	// Map: u = (1 + y) / (1 - y) mod p
	// p = 2^255 - 19

	// y is the input bytes with the high bit masked out (little endian)
	var yBytes [32]byte
	copy(yBytes[:], edPub[:])
	yBytes[31] &= 0x7F // clear sign bit

	// Use math/big for the field arithmetic
	y := new(big.Int).SetBytes(reverse(yBytes[:])) // big.Int is BigEndian, bytes are LittleEndian
	one := big.NewInt(1)

	// P = 2^255 - 19
	p, _ := new(big.Int).SetString("7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed", 16)

	// num = 1 + y
	num := new(big.Int).Add(one, y)

	// den = 1 - y
	den := new(big.Int).Sub(one, y)
	den.Mod(den, p) // handle negative result

	// invDen = (1 - y)^-1
	invDen := new(big.Int).ModInverse(den, p)

	// u = num * invDen mod p
	u := new(big.Int).Mul(num, invDen)
	u.Mod(u, p)

	// Convert u back to 32 bytes little-endian
	uBytesBuff := u.Bytes()
	var curvePub [32]byte
	// Pad if needed? u.Bytes() returns minimal bytes
	// We need to reverse back to Little Endian and put into 32 bytes
	for i := 0; i < len(uBytesBuff); i++ {
		// uBytesBuff is Big Endian.
		// curvePub needs Little Endian.
		// Last byte of uBytesBuff is the LSByte.
		// curvePub[0] is LSByte.
		curvePub[len(uBytesBuff)-1-i] = uBytesBuff[i]
	}
	// Note: If uBytesBuff is shorter than 32, the higher indices of curvePub (which map to lower indices of u) will be 0?
	// Wait, loop above:
	// Example u=1. uBytesBuff=[1]. len=1.
	// curvePub[0] = 1. Correct.

	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return "", "", err
	}

	msgBytes := []byte(message)
	ciphertext := box.Seal(nil, msgBytes, &nonce, &curvePub, &a.boxPrivKey)

	return hex.EncodeToString(ciphertext), hex.EncodeToString(nonce[:]), nil
}

// Helper to reverse bytes for BigInt conversion
func reverse(b []byte) []byte {
	r := make([]byte, len(b))
	for i := 0; i < len(b); i++ {
		r[i] = b[len(b)-1-i]
	}
	return r
}

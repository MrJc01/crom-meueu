# Crom Protocol Specification (v1)

This document defines the cryptographic standards used by the Crom Protocol to ensure compatibility between different client implementations (JS, Go, Rust, etc.).

## 1. Identity (Ed25519)
The core identity is an **Ed25519 KeyPair**.

### Key Derivation
To allow "Brain Keys" (Seed Phrases), we strictly follow this derivation path:
1. **Input**: Seed Phrase (String)
2. **Hash**: `SHA-256(seed_phrase)` -> 32 bytes
3. **Key Generation**: Use the 32-byte hash as the **Seed** for `ed25519.NewKeyFromSeed()`.

> **Note**: This produces a standard Ed25519 Private Key (64 bytes) and Public Key (32 bytes).

## 2. Encryption (X25519 / NaCl Box)
Direct Messages (DMs) usage **X25519 (Curve25519)** keys. Since users only have Ed25519 keys, we **derive** encryption keys mathematically.

### Ed25519 to Curve25519 Conversion
Compatible with `ed2curve.js` and `agl/extra25519`.

**Private Key Conversion:**
1. Take the first 32 bytes of the Ed25519 Private Key (the seed).
2. Compute `SHA-512(seed)`.
3. Clamp the first 32 bytes of the hash:
   - `digest[0] &= 248`
   - `digest[31] &= 127`
   - `digest[31] |= 64`
4. The result is the X25519 Private Scalar.

**Public Key Conversion:**
- Ed25519 Points (y-coordinate) must be converted to Montgomery form (u-coordinate).
- Formula: `u = (1 + y) / (1 - y)` (Birational map).
- Note: This is complex to implement from scratch. Use established libraries (`ed2curve`, `extra25519`).

### Message Format
DMs are encrypted using **NaCl Box** (xsalsa20-poly1305).
- **Nonce**: 24 bytes (Random)
- **Ciphertext**: `box(message, nonce, recipientCurvePub, senderCurvePriv)`

## 3. Signing & Canonicalization
To prevent signature manipulation, payloads are canonicalized before signing.

### format
String format:
```
version:1|author:<HEX_PUBKEY>|kind:<KIND>|timestamp:<UNIX_SEC>|payload:<JSON_STRING>
```

- **JSON String**: Must be standard JSON encoding of the payload object.
- **Signature**: `Ed25519_Sign(privKey, canonical_string)` -> Hex String.

## 4. Node Structure
A published node looks like this:

```json
{
  "id": "generated-by-server",
  "author_pubkey": "hex...",
  "kind": "text",
  "payload": { "content": "Hello" },
  "tags": ["recipient:hex..."],
  "signature": "hex...",
  "claimed_at": "ISO-8601"
}
```

## 5. Implementations
- **JavaScript**: Uses `nacl.min.js` + `ed2curve.min.js`.
- **Go**: Uses `crypto/ed25519` + `github.com/agl/ed25519/extra25519`.

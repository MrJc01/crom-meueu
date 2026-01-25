# Crom Protocol - Frontend Integration Guide

This guide explains how to interact with the Crom Protocol (MEUEU Node) using the Frontend SDK or by implementing your own client.

## 1. Cryptographic Identity (CromAuth)

The protocol uses **Ed25519** for signing and **X25519** for encryption.

### Key Generation
1. **Seed**: Users provide a "Seed Phrase" (Brain Key).
2. **Hash**: The seed is hashed using `SHA-256`.
3. **KeyPair**: Generate an Ed25519 KeyPair from the SHA-256 hash.

```javascript
// Example using noble-ed25519 or nacl
const hash = sha256(seedPhrase);
const keyPair = nacl.sign.keyPair.fromSeed(hash);
```

### Encryption Key Derivation (NaCl Box)
Direct Messages use ECDH (X25519). You must convert Ed25519 keys to X25519 (Curve25519) keys.
- **Tools**: `ed2curve.js` or `libsodium`.
- **Public Key**: `ed2curve.convertPublicKey(ed25519Pk)`
- **Secret Key**: `ed2curve.convertSecretKey(ed25519Sk)`

## 2. Signing & Publishing

To post content, you must sign a **Canonical String** of the payload.

### Canonical Format
```
version:1|author:<PUBKEY_HEX>|kind:<KIND>|timestamp:<UNIX_INT>|payload:<JSON_STRING>
```

**Parametros:**
- `PUBKEY_HEX`: Your Ed25519 public key in hex.
- `KIND`: String type (e.g., "post", "dm", "video").
- `UNIX_INT`: Integer timestamp (seconds).
- `JSON_STRING`: The raw JSON string of the payload object.

**Steps:**
1. Construct the string.
2. Sign using Ed25519 Secret Key.
3. Attach signature to the API request.

### API Endpoint: `POST /v1/publish`
Body:
```json
{
  "author_pubkey": "...",
  "kind": "post",
  "payload": { "content": "Hello World" },
  "claimed_at": "2024-01-01T00:00:00Z",
  "signature": "HexSignature..."
}
```

## 3. Direct Messages (Encrypted)

DMs are just nodes with `kind: "dm"`. The `payload` contains the ciphertext.

**Encryption Flow:**
1. Generate ephemeral nonce (24 bytes).
2. Encrypt message using NaCl Box (`curve25519-xsalsa20-poly1305`).
   - `Sender Secret Key (converted)` + `Recipient Public Key (converted)`.
3. Payload:
```json
{
  "ciphertext": "hex...",
  "nonce": "hex..."
}
```

## 4. Admin / Whitelist

If the server is in **WHITELIST MODE**:
- You can only publish if your Public Key is in the `whitelist` table.
- Contact the node admin to get approved.
- Use the Admin Dashboard at `/frontend/admin.html` if you are the operator.

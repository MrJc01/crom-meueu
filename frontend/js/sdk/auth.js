/**
 * Crom Protocol Auth SDK
 * Handles Identity, Signing, and Encryption.
 * Dependencies: nacl.min.js, ed2curve.min.js
 */
class CromAuth {
    constructor() {
        this.keyPair = null; // Ed25519
        this.boxKeyPair = null; // X25519
        this.pubKeyHex = null;
    }

    /**
     * Generate identity from a seed phrase (Brain Key)
     * @param {string} seedPhrase - The secret passphrase
     * @returns {Promise<string>} The Public Key (Hex)
     */
    async login(seedPhrase) {
        // 1. Hash Seed -> 32 bytes
        const encoder = new TextEncoder();
        const seedBytes = encoder.encode(seedPhrase);
        const hashBuffer = await crypto.subtle.digest('SHA-256', seedBytes);
        const hashArray = new Uint8Array(hashBuffer);

        // 2. Generate Ed25519 KeyPair (Signing)
        this.keyPair = nacl.sign.keyPair.fromSeed(hashArray);
        this.pubKeyHex = this.toHex(this.keyPair.publicKey);

        // 3. Derive X25519 KeyPair (Encryption)
        this.derivedEncryptionKeys();

        return this.pubKeyHex;
    }

    derivedEncryptionKeys() {
        if (typeof ed2curve === 'undefined') throw new Error("ed2curve library missing");

        const x25519Secret = ed2curve.convertSecretKey(this.keyPair.secretKey);
        const x25519Public = ed2curve.convertPublicKey(this.keyPair.publicKey);

        this.boxKeyPair = {
            publicKey: x25519Public,
            secretKey: x25519Secret
        };
    }

    /**
     * Sign a message (for Posting)
     * @param {string} message - The canonical message string
     * @returns {string} Hex signature
     */
    sign(message) {
        if (!this.keyPair) throw new Error("Not logged in");
        const msgBytes = new TextEncoder().encode(message);
        const signature = nacl.sign.detached(msgBytes, this.keyPair.secretKey);
        return this.toHex(signature);
    }

    /**
     * Encrypt a Direct Message
     * @param {string} recipientPubHex - Recipient's Ed25519 Public Key
     * @param {string} plaintext 
     * @returns {Object} { ciphertext: hex, nonce: hex }
     */
    encryptDM(recipientPubHex, plaintext) {
        if (!this.boxKeyPair) throw new Error("Not logged in");

        // Convert Recipient Ed25519 -> X25519
        const recipientEdPk = this.fromHex(recipientPubHex);
        const recipientCurvePk = ed2curve.convertPublicKey(recipientEdPk);
        if (!recipientCurvePk) throw new Error("Invalid Recipient Key");

        const nonce = nacl.randomBytes(nacl.box.nonceLength);
        const msgUint8 = new TextEncoder().encode(plaintext);

        const ciphertext = nacl.box(
            msgUint8,
            nonce,
            recipientCurvePk,
            this.boxKeyPair.secretKey
        );

        return {
            ciphertext: this.toHex(ciphertext),
            nonce: this.toHex(nonce)
        };
    }

    /**
     * Decrypt a Direct Message
     * @param {string} senderPubHex - Sender's Ed25519 Public Key
     * @param {string} ciphertextHex 
     * @param {string} nonceHex 
     * @returns {string|null} Decrypted text or null
     */
    decryptDM(senderPubHex, ciphertextHex, nonceHex) {
        if (!this.boxKeyPair) throw new Error("Not logged in");

        const senderEdPk = this.fromHex(senderPubHex);
        const senderCurvePk = ed2curve.convertPublicKey(senderEdPk);

        const ciphertext = this.fromHex(ciphertextHex);
        const nonce = this.fromHex(nonceHex);

        const decrypted = nacl.box.open(
            ciphertext,
            nonce,
            senderCurvePk,
            this.boxKeyPair.secretKey
        );

        if (!decrypted) return null;
        return new TextDecoder().decode(decrypted);
    }

    // Utils
    toHex(uint8Array) {
        return Array.from(uint8Array).map(b => b.toString(16).padStart(2, '0')).join('');
    }

    fromHex(hexString) {
        return new Uint8Array(hexString.match(/.{1,2}/g).map(byte => parseInt(byte, 16)));
    }
}

// Export global for now (legacy support)
window.CromAuth = CromAuth;

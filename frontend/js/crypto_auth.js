// IdentityManager - Handles Keys, Signing, and Encryption on the Client Side

class IdentityManager {
    constructor() {
        this.keyPair = null; // Ed25519 KeyPair
        this.boxKeyPair = null; // X25519 KeyPair (derived)
        this.pubKeyHex = null;
    }

    // Login (Seed Phrase / Brain Key)
    async login(seedPhrase) {
        // Deterministic Key Generation from SHA-256 hash of seed
        const encoder = new TextEncoder();
        const seedBytes = encoder.encode(seedPhrase);

        // Hash the seed to get 32 bytes
        const hashBuffer = await crypto.subtle.digest('SHA-256', seedBytes);
        const hashArray = new Uint8Array(hashBuffer);

        // Generate Ed25519 KeyPair
        this.keyPair = nacl.sign.keyPair.fromSeed(hashArray);
        this.pubKeyHex = this.toHex(this.keyPair.publicKey);

        // Derive Encryption Keys
        this.getEncryptionKeys();

        console.log(`🔑 Logged in as: ${this.pubKeyHex}`);
        sessionStorage.setItem('crom_identity', JSON.stringify({
            pubKey: this.pubKeyHex,
            seed: seedPhrase // In a real app, maybe don't store seed in plaintext session
        }));

        return this.pubKeyHex;
    }

    // Conversão Ed25519 -> X25519 via ed2curve
    getEncryptionKeys() {
        if (typeof ed2curve === 'undefined') {
            console.error("ed2curve library not loaded!");
            alert("Security Library Missing");
            throw new Error("ed2curve missing");
        }

        const x25519Secret = ed2curve.convertSecretKey(this.keyPair.secretKey);
        const x25519Public = ed2curve.convertPublicKey(this.keyPair.publicKey);

        if (!x25519Secret || !x25519Public) {
            throw new Error("Key conversion failed");
        }

        this.boxKeyPair = {
            publicKey: x25519Public,
            secretKey: x25519Secret
        };
    }

    // Auto-login from Session
    async checkSession() {
        const stored = JSON.parse(sessionStorage.getItem('crom_identity'));
        if (stored && stored.seed) {
            return await this.login(stored.seed);
        }
        return null;
    }

    logout() {
        this.keyPair = null;
        this.boxKeyPair = null;
        this.pubKeyHex = null;
        sessionStorage.removeItem('crom_identity');
        window.location.reload();
    }

    // Sign payload (for Public Post)
    sign(messageString) {
        if (!this.keyPair) throw new Error("Not logged in");

        const msgBytes = new TextEncoder().encode(messageString);
        const signature = nacl.sign.detached(msgBytes, this.keyPair.secretKey);

        return this.toHex(signature);
    }

    // Encrypt Direct Message (Box Creation)
    encryptDirectMessage(recipientPubHex, textContent) {
        if (!this.boxKeyPair) throw new Error("Not logged in or valid keys missing");

        // Convert User Public Key (Ed25519) -> X25519
        const recipientEdPk = this.fromHex(recipientPubHex);
        const recipientCurvePk = ed2curve.convertPublicKey(recipientEdPk);
        if (!recipientCurvePk) throw new Error("Invalid Recipient Key");

        const nonce = nacl.randomBytes(nacl.box.nonceLength);
        const msgUint8 = new TextEncoder().encode(textContent);

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

    // Decrypt Direct Message (Box Opening)
    decryptDirectMessage(senderPubHex, ciphertextHex, nonceHex) {
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

    toHex(uint8Array) {
        return Array.from(uint8Array)
            .map(b => b.toString(16).padStart(2, '0'))
            .join('');
    }

    fromHex(hexString) {
        return new Uint8Array(hexString.match(/.{1,2}/g).map(byte => parseInt(byte, 16)));
    }
}

// Global Instance
window.identityManager = new IdentityManager();

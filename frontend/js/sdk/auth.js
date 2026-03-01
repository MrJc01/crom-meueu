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
     * Generates a new identity, encrypts the private key with the given passphrase,
     * and returns the .cromid JSON object.
     */
    async generateIdentity(passphrase, trustedSource = '') {
        const randomBytes = new Uint8Array(32);
        crypto.getRandomValues(randomBytes);
        this.keyPair = nacl.sign.keyPair.fromSeed(randomBytes);
        this.pubKeyHex = this.toHex(this.keyPair.publicKey);
        this.derivedEncryptionKeys();

        const encryptedVault = await this.encryptSecretKey(this.keyPair.secretKey, passphrase);

        const cromid = {
            version: 1,
            pubKey: this.pubKeyHex,
            vault: encryptedVault
        };

        // Optional: User-defined trusted source for frontend integrity verification
        if (trustedSource) {
            cromid.trusted_source = trustedSource;
        }

        return cromid;
    }

    /**
     * Generates a new identity WITHOUT passphrase (v2 plain).
     * The secretKey is stored as hex in the .cromid file.
     * @returns {Object} The .cromid JSON object
     */
    generateIdentityPlain() {
        const randomBytes = new Uint8Array(32);
        crypto.getRandomValues(randomBytes);
        this.keyPair = nacl.sign.keyPair.fromSeed(randomBytes);
        this.pubKeyHex = this.toHex(this.keyPair.publicKey);
        this.derivedEncryptionKeys();

        return {
            version: 2,
            pubKey: this.pubKeyHex,
            secretKey: this.toHex(this.keyPair.secretKey)
        };
    }

    /**
     * Loads and decrypts an identity from a .cromid JSON object using a passphrase.
     */
    async loadIdentity(cromidData, passphrase) {
        if (cromidData.version !== 1) throw new Error("Unsupported .cromid version");

        const secretKey = await this.decryptSecretKey(cromidData.vault, passphrase);
        if (!secretKey) throw new Error("Invalid passphrase or corrupted file");

        this.keyPair = nacl.sign.keyPair.fromSecretKey(secretKey);
        this.pubKeyHex = this.toHex(this.keyPair.publicKey);

        if (this.pubKeyHex !== cromidData.pubKey) {
            throw new Error("Public key mismatch");
        }

        // Store trusted source for integrity checker
        this._trustedSource = cromidData.trusted_source || '';
        this._cromidData = cromidData;

        this.derivedEncryptionKeys();
        return this.pubKeyHex;
    }

    /**
     * Loads a v2 plain .cromid file (no passphrase needed).
     * Auto-detects version and calls the correct load method.
     */
    async loadIdentityAuto(cromidData) {
        if (cromidData.version === 2 && cromidData.secretKey) {
            // V2 plain — just load the secretKey directly
            const secretKey = this.fromHex(cromidData.secretKey);
            this.keyPair = nacl.sign.keyPair.fromSecretKey(secretKey);
            this.pubKeyHex = this.toHex(this.keyPair.publicKey);

            if (this.pubKeyHex !== cromidData.pubKey) {
                throw new Error("Public key mismatch");
            }

            this._trustedSource = cromidData.trusted_source || '';
            this._cromidData = cromidData;
            this.derivedEncryptionKeys();
            return this.pubKeyHex;
        } else if (cromidData.version === 1) {
            // V1 encrypted — needs passphrase (prompt user)
            const pass = prompt('This is an encrypted vault. Enter passphrase:');
            if (!pass) throw new Error('Passphrase required for v1 vault');
            return this.loadIdentity(cromidData, pass);
        } else {
            throw new Error('Unsupported .cromid version');
        }
    }

    /**
     * Export the .cromid vault as a downloadable JSON file.
     * @param {Object} cromidData - The .cromid JSON object to export
     */
    exportIdentity(cromidData) {
        const json = JSON.stringify(cromidData, null, 2);
        const blob = new Blob([json], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${cromidData.pubKey.substring(0, 8)}.cromid`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }

    /**
     * Returns the trusted source URL from the loaded .cromid, if any.
     * @returns {string} The trusted source URL or empty string
     */
    getTrustedSource() {
        return this._trustedSource || '';
    }

    // --- Vault Helper Functions ---

    async deriveKeyFromPassphrase(passphrase, salt) {
        const enc = new TextEncoder();
        const keyMaterial = await crypto.subtle.importKey(
            "raw",
            enc.encode(passphrase),
            { name: "PBKDF2" },
            false,
            ["deriveBits", "deriveKey"]
        );
        return crypto.subtle.deriveKey(
            {
                name: "PBKDF2",
                salt: salt,
                iterations: 100000,
                hash: "SHA-256"
            },
            keyMaterial,
            { name: "AES-GCM", length: 256 },
            true,
            ["encrypt", "decrypt"]
        );
    }

    async encryptSecretKey(secretKeyUint8, passphrase) {
        const salt = crypto.getRandomValues(new Uint8Array(16));
        const iv = crypto.getRandomValues(new Uint8Array(12));
        const key = await this.deriveKeyFromPassphrase(passphrase, salt);

        const encryptedContent = await crypto.subtle.encrypt(
            {
                name: "AES-GCM",
                iv: iv
            },
            key,
            secretKeyUint8
        );

        return {
            salt: this.toHex(salt),
            iv: this.toHex(iv),
            ciphertext: this.toHex(new Uint8Array(encryptedContent))
        };
    }

    async decryptSecretKey(vault, passphrase) {
        try {
            const salt = this.fromHex(vault.salt);
            const iv = this.fromHex(vault.iv);
            const ciphertext = this.fromHex(vault.ciphertext);
            const key = await this.deriveKeyFromPassphrase(passphrase, salt);

            const decryptedContent = await crypto.subtle.decrypt(
                {
                    name: "AES-GCM",
                    iv: iv
                },
                key,
                ciphertext
            );
            return new Uint8Array(decryptedContent);
        } catch (e) {
            return null;
        }
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

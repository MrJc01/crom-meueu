// Encrypted Direct Messages Logic

class DMManager {
    get im() {
        return window.cromAuth || window.identityManager;
    }

    // Send an Encrypted DM
    async sendDM(recipientPubHex, textContent) {
        if (!this.im.keyPair) {
            alert("Please login first!");
            return;
        }

        try {
            // New High-Level API from IdentityManager
            const encryptedData = this.im.encryptDM(recipientPubHex, textContent);

            // 2. Wrap in JSON Payload
            const encryptedPayload = {
                ciphertext: encryptedData.ciphertext,
                nonce: encryptedData.nonce,
                recipient_pubkey: recipientPubHex // Store so we know who it's for context
            };

            // 3. Publish to API with "recipient:KEY" tag
            const signedPayload = await this.buildSignedPayload(encryptedPayload, ["recipient:" + recipientPubHex]);

            const baseUrl = (window.serverManager ? window.serverManager.currentServer : window.location.origin);
            const resp = await fetch(baseUrl + '/v1/publish', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(signedPayload)
            });

            if (resp.status === 201) {
                // alert("🔒 Encrypted Message Sent!"); // Too intrusive
                return true;
            } else {
                alert("Failed to send: " + resp.status);
                return false;
            }

        } catch (e) {
            console.error(e);
            alert("Encryption Failed: " + e.message);
            return false;
        }
    }

    async buildSignedPayload(payloadObj, tags = []) {
        // Reuse IdentityManager signing logic if possible, but we need to construct the canonical string
        const timestamp = Math.floor(Date.now() / 1000);
        const kind = "encrypted_dm";
        const authorHex = this.im.pubKeyHex;

        const canonical = `version:1|author:${authorHex}|kind:${kind}|timestamp:${timestamp}|payload:${JSON.stringify(payloadObj)}`;

        // We use im.sign, but im.sign takes a messageString. 
        // Wait, im.sign() returns HEX signature.
        // We need to pass the canonical string to it.
        // Wait, im.sign in crypto_auth.js implementation:
        // sign(messageString) { ... const signature = nacl.sign.detached(...) ... return toHex(signature) }
        // Yes, that works.

        const sigHex = this.im.sign(canonical);

        return {
            author_pubkey: authorHex,
            kind: kind,
            payload: payloadObj,
            tags: tags,
            signature: sigHex,
            claimed_at: new Date().toISOString()
        };
    }

    // Fetch and Decrypt Inbox
    async fetchInbox() {
        if (!this.im.keyPair) return [];

        const myTag = "recipient:" + this.im.pubKeyHex;

        // Use dynamic server URL
        const baseUrl = (window.serverManager ? window.serverManager.currentServer : window.location.origin);
        const response = await fetch(baseUrl + '/v1/query', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                filters: {
                    kinds: ["encrypted_dm"],
                    tags: [myTag]
                },
                limit: 50
            })
        });

        const nodes = await response.json();
        return nodes.map(node => this.tryDecrypt(node)).filter(n => n !== null);
    }

    tryDecrypt(node) {
        try {
            const payload = node.payload;

            const content = this.im.decryptDM(
                node.author_pubkey,
                payload.ciphertext,
                payload.nonce
            );

            if (!content) return null;

            return {
                id: node.id,
                sender: node.author_pubkey,
                sent_at: node.claimed_at,
                content: content
            };

        } catch (e) {
            console.error("Decryption fail for node " + node.id, e);
            return null;
        }
    }
}

window.dmManager = new DMManager();

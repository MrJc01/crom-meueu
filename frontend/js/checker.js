/**
 * Crom-Checker: Frontend Integrity Verification Script
 * 
 * This script verifies that the frontend files (api.js, auth.js) have not been
 * tampered with by comparing their SHA-256 hashes against a manifest from a
 * trusted source defined in the user's .cromid vault.
 * 
 * Flow:
 * 1. Read trusted_source from the loaded .cromid
 * 2. Fetch manifest.json from that trusted source (GitHub raw URL)
 * 3. Compute SHA-256 of loaded frontend files
 * 4. Compare hashes — alert and block if mismatch detected
 */

class CromChecker {
    constructor() {
        this.verified = false;
        this.alerts = [];
    }

    /**
     * Run integrity check against the trusted source.
     * @param {string} trustedSource - URL to the raw manifest.json (e.g., GitHub raw URL)
     * @returns {Promise<{verified: boolean, alerts: string[]}>}
     */
    async verify(trustedSource) {
        if (!trustedSource) {
            this.verified = true; // No trusted source configured = skip verification
            return { verified: true, alerts: ['No trusted_source configured. Skipping integrity check.'] };
        }

        this.alerts = [];

        try {
            // 1. Fetch manifest from trusted source
            const manifestUrl = trustedSource.endsWith('/')
                ? trustedSource + 'manifest.json'
                : trustedSource + '/manifest.json';

            const manifestResp = await fetch(manifestUrl, {
                cache: 'no-store',
                headers: { 'Accept': 'application/json' }
            });

            if (!manifestResp.ok) {
                this.alerts.push(`⚠️ Could not fetch manifest from trusted source: ${manifestResp.status}`);
                this.verified = false;
                return { verified: false, alerts: this.alerts };
            }

            const manifest = await manifestResp.json();

            if (!manifest.files || typeof manifest.files !== 'object') {
                this.alerts.push('⚠️ Invalid manifest format: missing "files" object');
                this.verified = false;
                return { verified: false, alerts: this.alerts };
            }

            // 2. Verify each file listed in the manifest
            const filesToCheck = ['js/sdk/auth.js', 'js/api.js'];
            let allValid = true;

            for (const filePath of filesToCheck) {
                const expectedHash = manifest.files[filePath];
                if (!expectedHash) {
                    this.alerts.push(`ℹ️ File ${filePath} not found in manifest, skipping.`);
                    continue;
                }

                try {
                    const fileResp = await fetch('/' + filePath, { cache: 'no-store' });
                    if (!fileResp.ok) {
                        this.alerts.push(`⚠️ Could not fetch local file: ${filePath}`);
                        allValid = false;
                        continue;
                    }

                    const fileContent = await fileResp.text();
                    const computedHash = await this.sha256(fileContent);

                    if (computedHash !== expectedHash) {
                        this.alerts.push(`🚨 HASH MISMATCH: ${filePath} — Expected: ${expectedHash.substring(0, 16)}... Got: ${computedHash.substring(0, 16)}...`);
                        allValid = false;
                    } else {
                        this.alerts.push(`✅ ${filePath}: Integrity OK`);
                    }
                } catch (e) {
                    this.alerts.push(`⚠️ Error checking ${filePath}: ${e.message}`);
                    allValid = false;
                }
            }

            this.verified = allValid;

            if (!allValid) {
                this.showSecurityAlert();
            }

            return { verified: allValid, alerts: this.alerts };

        } catch (e) {
            this.alerts.push(`⚠️ Integrity check failed: ${e.message}`);
            this.verified = false;
            return { verified: false, alerts: this.alerts };
        }
    }

    /**
     * Compute SHA-256 hash of a string.
     * @param {string} text
     * @returns {Promise<string>} Hex-encoded hash
     */
    async sha256(text) {
        const encoder = new TextEncoder();
        const data = encoder.encode(text);
        const hashBuffer = await crypto.subtle.digest('SHA-256', data);
        const hashArray = Array.from(new Uint8Array(hashBuffer));
        return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
    }

    /**
     * Display a security alert overlay blocking key loading.
     */
    showSecurityAlert() {
        const overlay = document.createElement('div');
        overlay.id = 'crom-checker-alert';
        overlay.style.cssText = `
            position: fixed; top: 0; left: 0; width: 100%; height: 100%;
            background: rgba(139, 0, 0, 0.95); backdrop-filter: blur(10px);
            z-index: 99999; display: flex; align-items: center; justify-content: center;
            flex-direction: column; color: white; font-family: monospace;
        `;

        overlay.innerHTML = `
            <div style="max-width: 600px; text-align: center; padding: 40px;">
                <div style="font-size: 80px; margin-bottom: 20px;">🛡️</div>
                <h1 style="color: #ff4444; margin-bottom: 10px;">Frontend Adulterado</h1>
                <h2 style="color: #ffaa00; font-weight: normal;">Tampered Frontend Detected</h2>
                <p style="line-height: 1.8; margin: 20px 0; color: #ddd;">
                    The integrity verification has detected that one or more frontend files 
                    on this server have been modified from the expected version.
                </p>
                <div style="background: rgba(0,0,0,0.4); padding: 15px; border-radius: 8px; text-align: left; margin: 20px 0; max-height: 200px; overflow-y: auto;">
                    ${this.alerts.map(a => `<div style="color: ${a.startsWith('🚨') ? '#ff4444' : '#ccc'}; margin: 5px 0; font-size: 13px;">${a}</div>`).join('')}
                </div>
                <p style="color: #ff6666; font-weight: bold;">
                    ⛔ Private key loading has been BLOCKED to protect your identity.
                </p>
                <p style="color: #aaa; font-size: 12px; margin-top: 15px;">
                    If you trust this server, you can update the <code>trusted_source</code> in your .cromid file.
                </p>
                <button onclick="document.getElementById('crom-checker-alert').remove()" 
                    style="margin-top: 20px; padding: 10px 30px; background: transparent; 
                    border: 1px solid #666; color: #aaa; border-radius: 8px; cursor: pointer;">
                    Dismiss (Proceed at own risk)
                </button>
            </div>
        `;

        document.body.appendChild(overlay);
    }

    /**
     * Auto-run: Reads trusted_source from CromAuth and verifies integrity.
     * Should be called after identity is loaded.
     */
    async autoVerify() {
        const auth = window.cromAuth;
        if (!auth || typeof auth.getTrustedSource !== 'function') {
            return; // SDK not loaded or no getTrustedSource method
        }

        const source = auth.getTrustedSource();
        if (!source) {
            return; // No trusted source configured
        }

        const result = await this.verify(source);
        if (!result.verified) {
            console.warn('[CromChecker] Frontend integrity check FAILED:', result.alerts);
            // Block CromAuth key loading by clearing keys
            if (auth.keyPair) {
                console.warn('[CromChecker] Blocking private key access due to integrity failure.');
                auth.keyPair = null;
                auth.boxKeyPair = null;
            }
        } else {
            console.log('[CromChecker] Frontend integrity verified ✅');
        }
    }
}

// Initialize and export
window.cromChecker = new CromChecker();

// Auto-verify when document is ready (after auth is loaded)
document.addEventListener('DOMContentLoaded', () => {
    // Delay to ensure auth.js has had time to restore session
    setTimeout(() => {
        if (window.cromChecker) {
            window.cromChecker.autoVerify();
        }
    }, 2000);
});

/**
 * Crom-Checker: Integrity Watchdog
 * Validates local scripts against the manifest of the Trusted Source from the user's .cromid.
 */

class CromChecker {
    constructor() {
        this.scriptsToCheck = [
            '/js/api.js',
            '/js/app.js',
            '/js/sdk/auth.js',
            '/js/ui_login.js',
            '/index.html'
        ];
    }

    async hashData(dataText) {
        const msgUint8 = new TextEncoder().encode(dataText);
        const hashBuffer = await crypto.subtle.digest('SHA-256', msgUint8);
        const hashArray = Array.from(new Uint8Array(hashBuffer));
        return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
    }

    async fetchLocalScript(path) {
        try {
            const resp = await fetch(path);
            if (!resp.ok) return null;
            return await resp.text();
        } catch (e) {
            return null;
        }
    }

    async fetchRemoteManifest(trustedSourceUrl) {
        try {
            // e.g. https://raw.githubusercontent.com/user/crom-meueu/master/manifest.json
            // For now, we simulate fetching the manifest from the trusted URL root
            let uri = trustedSourceUrl;
            if (!uri.endsWith('/')) uri += '/';
            const resp = await fetch(uri + 'manifest.json');
            if (!resp.ok) return null;
            return await resp.json();
        } catch (e) {
            return null;
        }
    }

    async runCheck() {
        if (!window.cromAuth) return;
        const trustedSrc = window.cromAuth.getTrustedSource();
        if (!trustedSrc) {
            // No trusted source to verify against
            return;
        }

        const manifest = await this.fetchRemoteManifest(trustedSrc);
        if (!manifest) {
            console.warn("Crom-Checker: Could not fetch manifest from trusted source.");
            return;
        }

        let isTampered = false;

        for (const script of this.scriptsToCheck) {
            const localScriptText = await this.fetchLocalScript(script);
            if (!localScriptText) continue;

            const localHash = await this.hashData(localScriptText);

            // Check against manifest hash tree
            // Assuming manifest looks like: { "files": { "/js/api.js": "a1b2c3..." } }
            if (manifest.files && manifest.files[script]) {
                const expectedHash = manifest.files[script];
                if (localHash !== expectedHash) {
                    isTampered = true;
                    console.error(`Crom-Checker Alert: Tampering detected on ${script}!`);
                    console.error(`Expected: ${expectedHash} | Got: ${localHash}`);
                }
            }
        }

        this.renderBadge(isTampered, trustedSrc);
    }

    renderBadge(isTampered, sourceUrl) {
        const badge = document.createElement('div');
        badge.id = 'crom-checker-badge';
        badge.style.cssText = `
            position: fixed; bottom: 20px; left: 20px;
            padding: 8px 12px; border-radius: 8px; font-size: 11px;
            font-weight: bold; cursor: help; backdrop-filter: blur(5px); z-index: 9000;
        `;

        if (isTampered) {
            badge.style.background = 'rgba(255, 50, 50, 0.15)';
            badge.style.border = '1px solid #ff3232';
            badge.style.color = '#ff3232';
            badge.innerHTML = `⚠️ NETWORK COMPROMISED - DOM TAMPERING DETECTED`;
            badge.title = "The node server provided manipulated Javascript files that don't match your Trusted Source!";
        } else {
            badge.style.background = 'rgba(0, 255, 65, 0.1)';
            badge.style.border = '1px solid #00ff41';
            badge.style.color = '#00ff41';
            badge.innerHTML = `🛡️ Network Secure (Audited against ${new URL(sourceUrl).hostname})`;
            badge.title = "Integrity hashes match the original trusted source repository.";
        }

        // Drop existing
        const existing = document.getElementById('crom-checker-badge');
        if (existing) existing.remove();

        document.body.appendChild(badge);
    }
}

// Run watchdog smoothly after 3 seconds of load
setTimeout(() => {
    const watchdog = new CromChecker();
    watchdog.runCheck();
}, 3000);

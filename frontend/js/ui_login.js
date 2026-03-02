// Login UI Manager
// Uses SDK: CromAuth

// Initialize SDK immediately to avoid race conditions
if (!window.cromAuth) {
    window.cromAuth = new CromAuth();
}
window.identityManager = window.cromAuth; // Backwards compatibility

document.addEventListener('DOMContentLoaded', async () => {
    injectLoginModal();

    // Restore session from storage
    const storedVault = sessionStorage.getItem('crom_vault');
    if (storedVault) {
        try {
            const cromidData = JSON.parse(storedVault);
            await window.cromAuth.loadIdentityAuto(cromidData);
        } catch (e) {
            console.error("Session restore failed", e);
            sessionStorage.removeItem('crom_vault');
        }
    }

    // Always update UI
    updateAuthUI();
});

function injectLoginModal() {
    const div = document.createElement('div');
    div.id = 'login-overlay';
    div.style.cssText = `
        position: fixed; top: 0; left: 0; width: 100%; height: 100%;
        background: rgba(0,0,0,0.85); backdrop-filter: blur(10px);
        z-index: 9999; display: none; align-items: center; justify-content: center;
    `;

    div.innerHTML = `
        <div style="background:#1a1a1a; padding:30px; border-radius:16px; width:100%; max-width:400px; text-align:center; border:1px solid #333; box-shadow: 0 20px 50px rgba(0,0,0,0.5);">
            <h2 style="margin-top:0; color:white">Identity Vault</h2>
            <p style="color:#aaa; font-size:14px; margin-bottom:24px">Load your <b>.cromid</b> file or generate a new identity.</p>

            <!-- Login with file -->
            <label id="file-upload-label" style="
                display: flex; align-items: center; justify-content: center; gap: 10px;
                width: 100%; padding: 16px; border-radius: 12px;
                border: 2px dashed #444; background: #222; color: #ccc;
                font-size: 15px; font-weight: 500; cursor: pointer;
                transition: all 0.2s; margin-bottom: 12px;
            ">
                📂 Upload .cromid Vault
                <input type="file" id="cromid-file" accept=".cromid,.json" style="display:none">
            </label>

            <label id="restore-file-label" style="
                display: flex; align-items: center; justify-content: center; gap: 10px;
                width: 100%; padding: 12px; border-radius: 12px;
                background: rgba(255,255,255,0.05); color: #aaa; border: 1px solid #444;
                font-size: 13px; font-weight: 500; cursor: pointer;
                transition: all 0.2s; margin-bottom: 4px;
            ">
                📦 Import Backup (.zip)
                <input type="file" id="backup-file" accept=".zip" style="display:none">
            </label>

            <div style="display:flex; align-items:center; gap:12px; margin:20px 0;">
                <hr style="flex:1; border:0; border-top:1px solid #333;">
                <span style="color:#555; font-size:12px;">OR</span>
                <hr style="flex:1; border:0; border-top:1px solid #333;">
            </div>

            <!-- Generate new identity -->
            <button id="gen-btn" style="
                width: 100%; padding: 14px; border-radius: 12px;
                border: 1px solid #00ff41; background: rgba(0,255,65,0.08);
                color: #00ff41; font-size: 15px; font-weight: 600;
                cursor: pointer; transition: all 0.2s;
            ">⚡ Generate New Identity</button>
            <p style="color:#555; font-size:11px; margin-top:8px;">A <b>.cromid</b> file will be downloaded. Keep it safe!</p>

            <button id="cancel-login" style="margin-top:20px; background:none; border:none; color:#666; cursor:pointer; font-size:14px;">Cancel</button>
        </div>
    `;
    document.body.appendChild(div);

    // --- Event: Upload .cromid ---
    document.getElementById('cromid-file').onchange = async (e) => {
        const file = e.target.files[0];
        if (!file) return;

        const reader = new FileReader();
        reader.onload = async (ev) => {
            try {
                const cromidData = JSON.parse(ev.target.result);
                await window.cromAuth.loadIdentityAuto(cromidData);

                // Persist session
                sessionStorage.setItem('crom_vault', JSON.stringify(cromidData));
                div.style.display = 'none';
                updateAuthUI();
            } catch (err) {
                alert("Failed to load identity: " + err.message);
            }
        };
        reader.readAsText(file);
    };

    // --- Event: Restore Backup (.zip) ---
    const backupInput = document.getElementById('backup-file');
    if (backupInput) {
        backupInput.onchange = async (e) => {
            const file = e.target.files[0];
            if (!file) return;
            if (window.cromExporter) {
                await window.cromExporter.restoreBackup(file);
            } else {
                alert("Backup Exporter module not loaded.");
            }
        };
    }

    // Drag and Drop Effects
    const fileLabel = document.getElementById('file-upload-label');
    fileLabel.ondragover = (e) => {
        e.preventDefault();
        fileLabel.style.borderColor = '#00d2ff';
        fileLabel.style.background = 'rgba(0, 210, 255, 0.1)';
        fileLabel.style.color = '#fff';
    };
    fileLabel.ondragleave = (e) => {
        e.preventDefault();
        fileLabel.style.borderColor = '#444';
        fileLabel.style.background = '#222';
        fileLabel.style.color = '#ccc';
    };
    fileLabel.ondrop = (e) => {
        e.preventDefault();
        fileLabel.style.borderColor = '#444';
        fileLabel.style.background = '#222';
        if (e.dataTransfer.files.length) {
            document.getElementById('cromid-file').files = e.dataTransfer.files;
            document.getElementById('cromid-file').dispatchEvent(new Event('change'));
        }
    };
    fileLabel.onmouseenter = () => { fileLabel.style.borderColor = '#00d2ff'; fileLabel.style.color = '#fff'; };
    fileLabel.onmouseleave = () => { fileLabel.style.borderColor = '#444'; fileLabel.style.color = '#ccc'; };

    // --- Event: Generate New Identity ---
    document.getElementById('gen-btn').onclick = async () => {
        try {
            const pass = prompt("Create a Master Password for your Vault:");
            if (!pass) return;

            const cromidData = await window.cromAuth.generateIdentity(pass, window.location.origin);

            // Download the file
            window.cromAuth.exportIdentity(cromidData);

            // Persist session
            sessionStorage.setItem('crom_vault', JSON.stringify(cromidData));
            div.style.display = 'none';
            updateAuthUI();
        } catch (err) {
            alert("Failed to generate identity: " + err.message);
        }
    };

    // --- Cancel ---
    document.getElementById('cancel-login').onclick = () => {
        div.style.display = 'none';
    };
}

function showLogin() {
    document.getElementById('login-overlay').style.display = 'flex';
}

function updateAuthUI() {
    const auth = window.cromAuth;

    // Remove existing floating badge
    const existing = document.getElementById('auth-badge');
    if (existing) existing.remove();

    // Update sidebar identity (if present on page)
    const sidebarIdentity = document.getElementById('sidebar-identity');

    if (auth.pubKeyHex) {
        // ── LOGGED IN ──
        const hue = parseInt(auth.pubKeyHex.substring(0, 2), 16);

        // Update sidebar
        if (sidebarIdentity) {
            sidebarIdentity.innerHTML = `
                <div style="display:flex; align-items:center; gap:8px; padding:8px 12px; background:rgba(0,255,65,0.06); border:1px solid rgba(0,255,65,0.2); border-radius:8px; cursor:pointer; margin-bottom: 8px;" onclick="if(confirm('Logout?')){sessionStorage.removeItem('crom_identity');sessionStorage.removeItem('crom_vault');window.location.reload();}">
                    <span style="width:32px;height:32px;border-radius:50%;background:hsl(${hue},55%,45%);display:flex;align-items:center;justify-content:center;font-size:12px;font-weight:700;color:#fff;flex-shrink:0;">${auth.pubKeyHex.substring(0, 2).toUpperCase()}</span>
                    <div style="min-width:0;">
                        <div style="font-size:12px;font-weight:600;color:#4ade80;">● Online</div>
                        <div style="font-size:11px;color:#888;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">@${auth.pubKeyHex.substring(0, 12)}…</div>
                    </div>
                </div>
                <button id="export-backup-btn" onclick="window.cromExporter.exportBackup()" style="width:100%; padding:6px; background:rgba(0, 210, 255, 0.1); border:1px solid rgba(0,210,255,0.3); color:#00d2ff; font-size:11px; font-weight:bold; border-radius:6px; cursor:pointer; transition:0.2s">💾 Export Backup</button>
            `;
        }

        // Floating badge for pages without sidebar
        if (!sidebarIdentity) {
            const badge = document.createElement('div');
            badge.id = 'auth-badge';
            badge.style.cssText = `
                position: fixed; bottom: 90px; right: 20px;
                background: rgba(0,255,65,0.1); border: 1px solid #00ff41;
                color: #00ff41; padding: 8px 16px; border-radius: 20px;
                font-size: 12px; font-weight: bold; cursor: pointer;
                backdrop-filter: blur(5px); z-index: 1000;
            `;
            badge.innerHTML = `Ident: @${auth.pubKeyHex.substring(0, 6)}...`;
            badge.onclick = () => {
                if (confirm("Logout?")) {
                    sessionStorage.removeItem('crom_identity');
                    sessionStorage.removeItem('crom_vault');
                    window.location.reload();
                }
            };

            const expBadge = document.createElement('div');
            expBadge.id = 'export-badge';
            expBadge.style.cssText = `
                position: fixed; bottom: 130px; right: 20px;
                background: rgba(0,210,255,0.1); border: 1px solid #00d2ff;
                color: #00d2ff; padding: 6px 12px; border-radius: 20px;
                font-size: 11px; font-weight: bold; cursor: pointer;
                backdrop-filter: blur(5px); z-index: 1000;
            `;
            expBadge.innerHTML = `💾 Backup Data`;
            expBadge.onclick = () => {
                if (window.cromExporter) window.cromExporter.exportBackup();
            };

            document.body.appendChild(badge);
            document.body.appendChild(expBadge);
        }
    } else {
        // ── NOT LOGGED IN ──
        if (!sidebarIdentity) {
            const loginBtn = document.createElement('div');
            loginBtn.id = 'auth-badge';
            loginBtn.style.cssText = `
                position: fixed; bottom: 90px; right: 20px;
                background: #00d2ff; color: black;
                padding: 8px 16px; border-radius: 20px;
                font-size: 12px; font-weight: bold; cursor: pointer;
                box-shadow: 0 0 15px rgba(0, 210, 255, 0.4); z-index: 1000;
            `;
            loginBtn.innerText = "Login / Sign Up";
            loginBtn.onclick = showLogin;
            document.body.appendChild(loginBtn);
        }
    }
}

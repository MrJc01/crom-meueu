// Login UI Manager

document.addEventListener('DOMContentLoaded', async () => {
    injectLoginModal();

    // Auto-login check
    if (window.identityManager) {
        await window.identityManager.checkSession();
        updateAuthUI();
    }
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
            <h2 style="margin-top:0; color:white">Identity Login</h2>
            <p style="color:#aaa; font-size:14px; margin-bottom:20px">Enter your Brain Key (Seed Phrase) to access your decentralized identity.</p>
            
            <input type="password" id="seed-input" placeholder="Enter seed phrase..." style="width:100%; padding:12px; margin-bottom:15px; border-radius:8px; border:1px solid #444; background:#222; color:white; box-sizing:border-box">
            
            <button id="login-btn" style="width:100%; padding:12px; border-radius:8px; border:none; background:#00d2ff; color:black; font-weight:bold; cursor:pointer; margin-bottom:10px">Login</button>
            <button id="gen-btn" style="width:100%; padding:12px; border-radius:8px; border:1px solid #444; background:transparent; color:#fff; font-weight:bold; cursor:pointer">Generate New Identity</button>
            <button id="cancel-login" style="margin-top:15px; background:none; border:none; color:#666; cursor:pointer">Cancel</button>
        </div>
    `;
    document.body.appendChild(div);

    // Event Listeners
    document.getElementById('login-btn').onclick = async () => {
        const seed = document.getElementById('seed-input').value;
        if (!seed) return alert("Please enter a seed");
        await window.identityManager.login(seed);
        div.style.display = 'none';
        updateAuthUI();
    };

    document.getElementById('gen-btn').onclick = () => {
        // Generate random seed (simple version)
        const randomBytes = new Uint8Array(32);
        crypto.getRandomValues(randomBytes);
        const seed = Array.from(randomBytes).map(b => b.toString(16).padStart(2, '0')).join('');
        document.getElementById('seed-input').value = seed;
        document.getElementById('seed-input').type = "text"; // Show it
        alert("This is your NEW Identity Key. Save it somewhere safe!");
    };

    document.getElementById('cancel-login').onclick = () => {
        div.style.display = 'none';
    };
}

function showLogin() {
    document.getElementById('login-overlay').style.display = 'flex';
}

function updateAuthUI() {
    const im = window.identityManager;
    const nav = document.querySelector('.bottom-nav') || document.body;

    // Remove existing badge
    const existing = document.getElementById('auth-badge');
    if (existing) existing.remove();

    if (im.pubKeyHex) {
        const badge = document.createElement('div');
        badge.id = 'auth-badge';
        badge.style.cssText = `
            position: fixed; bottom: 90px; right: 20px;
            background: rgba(0, 255, 65, 0.1); border: 1px solid #00ff41;
            color: #00ff41; padding: 8px 16px; border-radius: 20px;
            font-size: 12px; font-weight: bold; cursor: pointer;
            backdrop-filter: blur(5px); z-index: 1000;
        `;
        badge.innerHTML = `Ident: @${im.pubKeyHex.substring(0, 6)}...`;
        badge.onclick = () => {
            if (confirm("Logout?")) im.logout();
        };
        document.body.appendChild(badge);
    } else {
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

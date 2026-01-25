// Server Switcher & Manager

class ServerManager {
    constructor() {
        this.defaultServer = "http://localhost:8080";
        this.currentServer = localStorage.getItem('crom_server') || this.defaultServer;
        console.log("🌍 Crom Network: Connected to " + this.currentServer);
    }

    setServer(url) {
        // Strip trailing slash
        url = url.replace(/\/$/, "");
        if (!url.startsWith("http")) {
            url = "http://" + url;
        }

        localStorage.setItem('crom_server', url);
        this.currentServer = url;
        window.location.reload(); // Refresh to apply
    }

    reset() {
        this.setServer(this.defaultServer);
    }

    // UI Injection
    injectUI() {
        const div = document.createElement('div');
        div.id = 'server-selector';
        div.style.cssText = `
            position: fixed; top: 10px; right: 10px;
            background: rgba(0,0,0,0.6); backdrop-filter: blur(5px);
            padding: 5px 10px; border-radius: 20px;
            border: 1px solid rgba(255,255,255,0.2);
            font-size: 12px; color: #aaa; cursor: pointer;
            z-index: 9999; display: flex; align-items: center; gap: 5px;
        `;

        const statusDot = this.currentServer.includes("localhost") ? "🟢" : "☁️";
        div.innerHTML = `${statusDot} ${this.currentServer.replace("http://", "")}`;

        div.onclick = () => this.showModal();
        document.body.appendChild(div);
    }

    showModal() {
        const modal = document.createElement('div');
        modal.style.cssText = `
            position: fixed; top: 0; left: 0; width: 100%; height: 100%;
            background: rgba(0,0,0,0.8); z-index: 10000;
            display: flex; align-items: center; justify-content: center;
        `;

        modal.innerHTML = `
            <div style="background:#222; padding:25px; border-radius:12px; width:300px; border:1px solid #444">
                <h3 style="margin-top:0; color:white">Select Server Node</h3>
                <p style="color:#888; font-size:13px">Choose which "Crom Node" you want to read/write to.</p>
                
                <div style="margin-bottom:15px">
                    <label style="color:white; display:block; margin-bottom:5px">Server URL:</label>
                    <input type="text" id="server-url-input" value="${this.currentServer}" style="width:100%; padding:8px; background:#111; border:1px solid #333; color:white; border-radius:4px">
                </div>
                
                <button id="save-server" style="width:100%; padding:10px; background:#00d2ff; border:none; border-radius:4px; font-weight:bold; cursor:pointer">Connect</button>
                <div style="margin-top:10px; display:flex; gap:10px">
                    <button id="preset-local" style="flex:1; padding:5px; background:#333; border:none; color:white; border-radius:4px; cursor:pointer">Localhost</button>
                    <button id="close-server" style="flex:1; padding:5px; background:none; border:1px solid #444; color:#888; border-radius:4px; cursor:pointer">Cancel</button>
                </div>
            </div>
        `;
        document.body.appendChild(modal);

        document.getElementById('save-server').onclick = () => {
            this.setServer(document.getElementById('server-url-input').value);
        };

        document.getElementById('preset-local').onclick = () => {
            this.setServer("http://localhost:8080");
        };

        document.getElementById('close-server').onclick = () => modal.remove();
    }
}

window.serverManager = new ServerManager();
document.addEventListener('DOMContentLoaded', () => window.serverManager.injectUI());

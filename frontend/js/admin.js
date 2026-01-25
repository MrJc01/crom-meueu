const AdminApp = {
    token: null,
    baseUrl: '/admin',

    init: () => {
        const storedToken = sessionStorage.getItem('admin_token');
        if (storedToken) {
            AdminApp.token = storedToken;
            AdminApp.showDashboard();
            AdminApp.loadAll();
        }
    },

    login: () => {
        const input = document.getElementById('admin-token');
        const token = input.value.trim();
        if (!token) return alert("Token required");

        AdminApp.token = token;
        // Verify by trying to fetch stats
        AdminApp.loadStats()
            .then(() => {
                sessionStorage.setItem('admin_token', token);
                AdminApp.showDashboard();
                AdminApp.loadAll();
            })
            .catch(() => {
                AdminApp.token = null;
                alert("Authentication Failed");
            });
    },

    logout: () => {
        sessionStorage.removeItem('admin_token');
        window.location.reload();
    },

    showDashboard: () => {
        document.getElementById('auth-section').classList.add('hidden');
        document.getElementById('dashboard-section').classList.remove('hidden');
        document.getElementById('connection-status').innerText = "STATUS: CONNECTED [ROOT]";
        document.getElementById('connection-status').style.color = "#00ff00";
    },

    headers: () => {
        return {
            'X-Admin-Token': AdminApp.token,
            'Content-Type': 'application/json'
        };
    },

    // API Calls
    loadAll: () => {
        AdminApp.loadStats();
        AdminApp.loadWhitelist();
        AdminApp.loadBannedWords();
    },

    loadStats: async () => {
        const res = await fetch(`${AdminApp.baseUrl}/stats`, { headers: AdminApp.headers() });
        if (!res.ok) throw new Error("Auth Failed");
        const data = await res.json();

        document.getElementById('stat-nodes').innerText = data.total_nodes;
        document.getElementById('stat-users').innerText = data.total_users;
        document.getElementById('stat-active').innerText = data.active_nodes_24h;
    },

    // Whitelist
    loadWhitelist: async () => {
        const res = await fetch(`${AdminApp.baseUrl}/whitelist`, { headers: AdminApp.headers() });
        const list = await res.json();
        const ul = document.getElementById('whitelist-list');
        ul.innerHTML = '';
        list.forEach(key => {
            const li = document.createElement('li');
            li.innerHTML = `<span>${key.substring(0, 16)}...</span> <button onclick="AdminApp.removeFromWhitelist('${key}')" class="danger">Revoke</button>`;
            ul.appendChild(li);
        });
    },

    addToWhitelist: async () => {
        const input = document.getElementById('whitelist-input');
        const key = input.value.trim();
        if (!key) return;

        await fetch(`${AdminApp.baseUrl}/whitelist`, {
            method: 'POST',
            headers: AdminApp.headers(),
            body: JSON.stringify({ public_key: key })
        });
        input.value = '';
        AdminApp.loadWhitelist();
    },

    removeFromWhitelist: async (key) => {
        if (!confirm("Revoke access?")) return;
        await fetch(`${AdminApp.baseUrl}/whitelist?public_key=${key}`, {
            method: 'DELETE',
            headers: AdminApp.headers()
        });
        AdminApp.loadWhitelist();
    },

    // Banned Words
    loadBannedWords: async () => {
        const res = await fetch(`${AdminApp.baseUrl}/banned_words`, { headers: AdminApp.headers() });
        const list = await res.json();
        const ul = document.getElementById('banword-list');
        ul.innerHTML = '';
        list.forEach(word => {
            const li = document.createElement('li');
            li.innerHTML = `<span>${word}</span> <button onclick="AdminApp.unbanWord('${word}')">Unban</button>`;
            ul.appendChild(li);
        });
    },

    banWord: async () => {
        const input = document.getElementById('banword-input');
        const word = input.value.trim();
        if (!word) return;

        await fetch(`${AdminApp.baseUrl}/banned_words`, {
            method: 'POST',
            headers: AdminApp.headers(),
            body: JSON.stringify({ word: word })
        });
        input.value = '';
        AdminApp.loadBannedWords();
    },

    unbanWord: async (word) => {
        await fetch(`${AdminApp.baseUrl}/banned_words?word=${word}`, {
            method: 'DELETE',
            headers: AdminApp.headers()
        });
        AdminApp.loadBannedWords();
    }
};

window.onload = AdminApp.init;

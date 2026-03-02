// API Manager
// Uses SDK: CromClient

const client = new CromClient(
    (window.serverManager ? window.serverManager.currentServer : '')
);
window.cromClient = client; // Global access

// Helper for wrappers
const API_URL = client.baseUrl + '/v1/query';

function getPublishURL() {
    return client.baseUrl + '/v1/publish';
}

async function fetchNodes(filters = {}) {
    // Use SDK
    try {
        return await client.query(filters);
    } catch (e) {
        console.error("SDK fetchNodes Error:", e);
        return [];
    }
}

function formatDate(isoString) {
    const d = new Date(isoString);
    return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

// Navigation Injection
document.addEventListener('DOMContentLoaded', () => {
    // Only inject if not already present
    if (document.querySelector('.bottom-nav')) return;

    // Toggle Button
    const toggleBtn = document.createElement('button');
    toggleBtn.innerHTML = '☰';
    toggleBtn.id = 'nav-toggle';
    toggleBtn.style.cssText = `
        position: fixed; bottom: 20px; right: 20px; z-index: 2000;
        width: 50px; height: 50px; border-radius: 50%;
        background: #00d2ff; color: #000; border: none;
        font-size: 24px; cursor: pointer;
        box-shadow: 0 5px 15px rgba(0,0,0,0.5);
    `;

    // Main Nav Container
    // Main Nav Container
    const nav = document.createElement('nav');
    nav.className = 'bottom-nav hidden'; // Start hidden
    nav.id = 'main-nav';
    // Style override to make it float correctly when shown
    nav.style.cssText = `
        display: flex; gap: 10px; align-items: center;
        position: fixed; bottom: 80px; right: 20px;
        background: rgba(15, 12, 41, 0.95); padding: 10px 20px;
        border-radius: 20px; border: 1px solid #333;
        backdrop-filter: blur(15px); z-index: 1999;
        transition: opacity 0.3s, transform 0.3s;
        transform: translateY(20px); opacity: 0; pointer-events: none;
        box-shadow: 0 10px 40px rgba(0,0,0,0.5);
        max-width: 90vw; overflow-x: auto; /* Responsive scroll */
    `;

    // Toggle Logic
    let isOpen = false;
    toggleBtn.onclick = () => {
        isOpen = !isOpen;
        if (isOpen) {
            nav.style.transform = 'translateY(0)';
            nav.style.opacity = '1';
            nav.style.pointerEvents = 'all';
            toggleBtn.innerHTML = '✖';
        } else {
            nav.style.transform = 'translateY(20px)';
            nav.style.opacity = '0';
            nav.style.pointerEvents = 'none';
            toggleBtn.innerHTML = '☰';
        }
    };

    if (document.body.classList.contains('tiktok-mode')) {
        nav.classList.add('tiktok-nav');
    }

    // Portal Button (distinct style)
    const portalBtn = document.createElement('a');
    portalBtn.href = '/';
    portalBtn.className = 'nav-item home-btn';
    portalBtn.innerHTML = '🪐';
    portalBtn.title = 'Portal';
    nav.appendChild(portalBtn);

    const separator = document.createElement('div');
    separator.style.width = '1px';
    separator.style.background = 'rgba(255,255,255,0.2)';
    separator.style.height = '30px';
    separator.style.margin = '0 10px';
    nav.appendChild(separator);

    const links = [
        { icon: '➕', href: 'publish.html', title: 'New' },
        { icon: '🌐', href: 'index.html', title: 'Feed' },
        { icon: '🔐', href: 'inbox.html', title: 'Inbox' },
        { icon: '🔍', href: 'explorer.html', title: 'Explorer' },
        { icon: '⚠️', href: 'admin.html', title: 'Admin' },
    ];

    const currentPath = window.location.pathname.split('/').pop() || 'index.html';

    links.forEach(link => {
        const a = document.createElement('a');
        a.href = link.href;
        a.className = 'nav-item';
        a.innerHTML = link.icon;
        a.title = link.title;
        if (currentPath === link.href) {
            a.classList.add('active');
        }
        nav.appendChild(a);
    });

    document.body.appendChild(nav);
    document.body.appendChild(toggleBtn);
});

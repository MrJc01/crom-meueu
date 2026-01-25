// API Base - Controlled by ServerManager
const API_URL = (window.serverManager ? window.serverManager.currentServer : '') + '/v1/query';

// Helper for Publish requests (used by DMManager etc)
function getPublishURL() {
    return (window.serverManager ? window.serverManager.currentServer : '') + '/v1/publish';
}

async function fetchNodes(filters = {}) {
    const url = (window.serverManager ? window.serverManager.currentServer : '') + '/v1/query';
    try {
        const response = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                filters: filters,
                limit: 50 // Increased limit for better feel
            })
        });

        if (!response.ok) {
            throw new Error(`API Error: ${response.status}`);
        }

        return await response.json();
    } catch (error) {
        console.error('Fetch error:', error);
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

    const nav = document.createElement('nav');
    nav.className = 'bottom-nav';
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
        { icon: '🐦', href: 'twitter.html', title: 'X' },
        { icon: '📱', href: 'tiktok.html', title: 'Tok' },
        { icon: '📺', href: 'youtube.html', title: 'Tube' },
        { icon: '📰', href: 'tabnews.html', title: 'News' },
        { icon: '👥', href: 'facebook.html', title: 'Face' },
        { icon: '🔐', href: 'inbox.html', title: 'Inbox' },
        { icon: '🔍', href: 'explorer.html', title: 'Explorer' },
    ];

    const currentPath = window.location.pathname.split('/').pop();

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
});

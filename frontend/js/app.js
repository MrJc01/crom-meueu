/**
 * Crom Protocol — Universal Feed App (Tailwind Edition)
 * Handles: API communication, dynamic rendering by kind, client-side filtering.
 */

const app = {
    currentFilter: 'all',
    nodes: [],

    init() {
        const baseUrl = window.serverManager ? window.serverManager.currentServer : '';
        if (!window.cromClient) {
            window.cromClient = new CromClient(baseUrl);
        }
        this.bindFilters();
        this.startInactivityTimer();
        this.loadFeed();
    },

    startInactivityTimer() {
        this.inactivityTimeoutListener = () => {
            clearTimeout(this.inactivityTimer);
            this.inactivityTimer = setTimeout(() => {
                if (window.cromAuth && window.cromAuth.pubKeyHex) {
                    console.log("Memory Wipe: Session expired due to inactivity.");
                    window.cromAuth.keyPair = null;
                    window.cromAuth.boxKeyPair = null;
                    window.cromAuth.pubKeyHex = null;
                    sessionStorage.removeItem('crom_vault');
                    alert("Session expired due to inactivity. Vault keys wiped from RAM.");
                    window.location.reload();
                }
            }, 15 * 60 * 1000); // 15 minutes
        };

        ['mousemove', 'keydown', 'scroll', 'click'].forEach(evt =>
            window.addEventListener(evt, this.inactivityTimeoutListener)
        );
        this.inactivityTimeoutListener(); // Init
    },

    bindFilters() {
        document.querySelectorAll('[data-filter]').forEach(btn => {
            btn.addEventListener('click', () => {
                this.currentFilter = btn.dataset.filter;
                this.renderFeed();
            });
        });
    },

    async loadFeed(filters = {}) {
        const container = document.getElementById('feed');
        container.innerHTML = `
            <div class="flex flex-col items-center gap-4 py-20 text-gray-500 text-sm">
                <div class="w-8 h-8 border-2 border-white/10 border-t-cyan-400 rounded-full spinner"></div>
                <span>Loading nodes…</span>
            </div>`;

        try {
            this.nodes = await window.cromClient.query(filters);
            this.renderFeed();
        } catch (e) {
            console.error('Feed load error:', e);
            container.innerHTML = `
                <div class="text-center py-20">
                    <span class="text-4xl mb-3 block">📡</span>
                    <p class="text-gray-400 text-sm">Could not connect to node</p>
                    <button onclick="app.loadFeed()" class="mt-4 px-4 py-2 rounded-lg bg-white/5 border border-white/10 text-xs text-gray-400 hover:text-white hover:bg-white/10 transition">Retry</button>
                </div>`;
        }
    },

    renderFeed() {
        const container = document.getElementById('feed');
        let nodes = this.nodes;

        if (this.currentFilter !== 'all') {
            const filterMap = {
                'text': ['text', 'text/article', 'note', 'comment', 'article'],
                'video': ['video', 'video/external', 'video/embed'],
                'image': ['image', 'image/url', 'photo'],
                'link': ['link', 'bookmark'],
            };
            const kinds = filterMap[this.currentFilter] || [this.currentFilter];
            nodes = nodes.filter(n => kinds.some(k => n.kind && n.kind.startsWith(k)));
        }

        if (!nodes || nodes.length === 0) {
            container.innerHTML = `
                <div class="text-center py-16 text-gray-500 text-sm">
                    <span class="text-3xl mb-2 block">🕳️</span>
                    No nodes found for this filter
                </div>`;
            return;
        }

        container.innerHTML = '<div class="space-y-4">' + nodes.map(node => this.renderNode(node)).join('') + '</div>';
    },

    renderNode(node) {
        const authorShort = (node.author_pubkey || '').substring(0, 8);
        const hue = parseInt((node.author_pubkey || '00').substring(0, 2), 16);
        const time = this.formatDate(node.claimed_at);
        const kind = node.kind || 'unknown';
        const payload = node.payload || {};

        let content = '';
        let mediaSection = '';
        const [badgeIcon, badgeLabel, badgeColor] = this.getKindInfo(kind);

        switch (kind) {
            case 'text/short':
                // Renderiza como um 'tweet' - micro-blog style
                content = `<p class="text-[15px] text-gray-200 leading-snug">${this.escapeHtml(payload.content || payload.text || '')}</p>`;
                break;
            case 'text/article':
                // Renderiza como um post de blog (Tabnews)
                content = `<h2 class="text-xl font-bold text-white mb-2">${this.escapeHtml(payload.title || 'Untitled Article')}</h2>`;
                content += `<div class="prose prose-invert max-w-none prose-sm mt-2"><p class="text-gray-300 leading-relaxed whitespace-pre-wrap">${this.escapeHtml(payload.content || payload.body || '')}</p></div>`;
                break;
            case 'video/mp4':
                // Renderiza um reprodutor de vídeo HTML5
                const vidUrl = payload.url || payload.video_url || '';
                if (vidUrl) {
                    mediaSection = `<div class="rounded-xl overflow-hidden my-3 bg-black border border-white/5"><video src="${this.escapeHtml(vidUrl)}" controls class="w-full max-h-[400px]" preload="metadata"></video></div>`;
                }
                const vidTitle = payload.title || payload.caption || '';
                if (vidTitle) {
                    content = `<p class="text-gray-300 mt-2 font-medium">${this.escapeHtml(vidTitle)}</p>`;
                }
                break;
            default:
                // Fallback for other kinds backward compatibility
                if (kind.startsWith('video') || kind === 'video/external') {
                    const embedUrl = this.getEmbedUrl(payload.url || payload.video_url || '');
                    if (embedUrl) {
                        mediaSection = `<div class="rounded-xl overflow-hidden my-3 bg-black"><iframe src="${embedUrl}" class="w-full aspect-video" frameborder="0" allowfullscreen loading="lazy"></iframe></div>`;
                    }
                    content = `<p class="text-gray-300">${this.escapeHtml(payload.title || payload.text || payload.content || '')}</p>`;
                } else if (kind.startsWith('image') || kind === 'photo') {
                    const imgUrl = payload.url || payload.image_url || '';
                    if (imgUrl) {
                        mediaSection = `<div class="rounded-xl overflow-hidden my-3"><img src="${this.escapeHtml(imgUrl)}" alt="Image" class="w-full max-h-[400px] object-cover" loading="lazy" onerror="this.style.display='none'"></div>`;
                    }
                    content = `<p class="text-gray-300">${this.escapeHtml(payload.caption || payload.text || payload.content || '')}</p>`;
                } else {
                    let baseContent = this.escapeHtml(payload.content || payload.text || payload.body || '');
                    if (payload.title) {
                        content = `<h3 class="font-semibold text-white text-lg mb-1">${this.escapeHtml(payload.title)}</h3><p class="text-gray-300 leading-relaxed">${baseContent}</p>`;
                    } else {
                        content = `<p class="text-gray-300 leading-relaxed">${baseContent}</p>`;
                    }
                }
                break;
        }

        if (payload.url && kind !== 'video/mp4' && !kind.startsWith('video') && !kind.startsWith('image')) {
            content += `<div class="mt-3"><a href="${this.escapeHtml(payload.url)}" target="_blank" rel="noopener" class="text-crom hover:underline text-xs truncate block max-w-full"><span class="bg-crom/10 px-2 py-1.5 rounded-lg inline-flex items-center gap-1.5 border border-crom/20 font-medium">🔗 ${this.escapeHtml(payload.url)}</span></a></div>`;
        }

        return `
        <article class="node-card bg-surface-800/80 backdrop-blur border border-white/[0.08] shadow-2xl rounded-2xl p-5 mb-4 hover:border-white/20 hover:bg-surface-800 transition-all group" data-kind="${this.escapeHtml(kind)}">
            <div class="flex items-center gap-3.5 mb-4">
                <a href="profile.html?pubkey=${this.escapeHtml(node.author_pubkey)}" class="w-11 h-11 rounded-full flex items-center justify-center text-sm font-bold text-white flex-shrink-0 shadow-lg group-hover:ring-2 ring-white/30 transition-all" style="background: hsl(${hue}, 60%, 45%)">
                    ${this.escapeHtml(authorShort.substring(0, 2).toUpperCase())}
                </a>
                <div class="flex-1 min-w-0">
                    <a href="profile.html?pubkey=${this.escapeHtml(node.author_pubkey)}" class="text-[15px] font-bold text-white hover:text-crom transition-colors">${this.escapeHtml(authorShort)}…</a>
                    <div class="text-xs text-gray-400 font-medium mt-0.5">${this.escapeHtml(time)}</div>
                </div>
                <span class="px-2.5 py-1 rounded-md text-[10px] uppercase tracking-widest font-bold shadow-sm ${badgeColor}">${badgeIcon} ${badgeLabel}</span>
            </div>
            ${mediaSection}
            <div class="pt-1 pb-2 ${content ? '' : 'text-gray-600 italic'}">${content || 'Empty payload'}</div>
            <div class="flex items-center gap-4 mt-3 pt-4 border-t border-white/[0.06] text-xs font-semibold">
                <a href="thread.html?id=${this.escapeHtml(node.id)}" class="text-gray-400 hover:text-crom transition flex items-center gap-1.5 bg-white/5 hover:bg-white/10 px-3 py-1.5 rounded-lg"><span class="text-sm">💬</span> Discuss</a>
                <span class="text-gray-600 font-mono ml-auto opacity-40 hover:opacity-100 transition cursor-help" title="${this.escapeHtml(node.id)}">#${(node.id || '').substring(0, 8)}</span>
            </div>
        </article>`;
    },

    getKindInfo(kind) {
        const map = {
            'text': ['📝', 'Text', 'bg-blue-500/15 text-blue-400 border border-blue-500/20'],
            'text/short': ['🐦', 'Post', 'bg-blue-500/15 text-blue-400 border border-blue-500/20'],
            'text/article': ['📰', 'Article', 'bg-purple-500/15 text-purple-400 border border-purple-500/20'],
            'article': ['📰', 'Article', 'bg-purple-500/15 text-purple-400 border border-purple-500/20'],
            'note': ['📝', 'Note', 'bg-blue-500/15 text-blue-400 border border-blue-500/20'],
            'comment': ['💬', 'Comment', 'bg-gray-500/15 text-gray-400 border border-gray-500/20'],
            'video': ['📹', 'Video', 'bg-red-500/15 text-red-400 border border-red-500/20'],
            'video/mp4': ['📹', 'HTML5', 'bg-red-500/15 text-red-400 border border-red-500/20'],
            'video/external': ['📺', 'Video', 'bg-red-500/15 text-red-400 border border-red-500/20'],
            'video/embed': ['📺', 'Embed', 'bg-red-500/15 text-red-400 border border-red-500/20'],
            'image': ['📸', 'Image', 'bg-pink-500/15 text-pink-400 border border-pink-500/20'],
            'image/url': ['📸', 'Image', 'bg-pink-500/15 text-pink-400 border border-pink-500/20'],
            'photo': ['📸', 'Photo', 'bg-pink-500/15 text-pink-400 border border-pink-500/20'],
            'link': ['🔗', 'Link', 'bg-yellow-500/15 text-yellow-400 border border-yellow-500/20'],
            'bookmark': ['🔖', 'Bookmark', 'bg-yellow-500/15 text-yellow-400 border border-yellow-500/20'],
        };
        return map[kind] || ['📦', kind, 'bg-gray-500/15 text-gray-400 border border-gray-500/20'];
    },

    getEmbedUrl(url) {
        if (!url) return null;
        const ytMatch = url.match(/(?:youtube\.com\/watch\?v=|youtu\.be\/)([^&\s]+)/);
        if (ytMatch) return `https://www.youtube.com/embed/${ytMatch[1]}`;
        const vimeoMatch = url.match(/vimeo\.com\/(\d+)/);
        if (vimeoMatch) return `https://player.vimeo.com/video/${vimeoMatch[1]}`;
        return null;
    },

    formatDate(isoString) {
        if (!isoString) return '';
        const d = new Date(isoString);
        const diff = Date.now() - d;
        if (diff < 60000) return 'just now';
        if (diff < 3600000) return `${Math.floor(diff / 60000)}m`;
        if (diff < 86400000) return `${Math.floor(diff / 3600000)}h`;
        return d.toLocaleDateString();
    },

    escapeHtml(str) {
        if (!str) return '';
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }
};

document.addEventListener('DOMContentLoaded', () => app.init());

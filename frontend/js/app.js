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
        this.loadFeed();
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

        if (kind.startsWith('video') || kind === 'video/external') {
            const videoUrl = payload.url || payload.video_url || '';
            const embedUrl = this.getEmbedUrl(videoUrl);
            if (embedUrl) {
                mediaSection = `<div class="rounded-xl overflow-hidden my-3 bg-black"><iframe src="${embedUrl}" class="w-full aspect-video" frameborder="0" allowfullscreen loading="lazy"></iframe></div>`;
            }
            content = payload.title || payload.text || payload.content || '';
        } else if (kind.startsWith('image') || kind === 'photo') {
            const imgUrl = payload.url || payload.image_url || '';
            if (imgUrl) {
                mediaSection = `<div class="rounded-xl overflow-hidden my-3"><img src="${imgUrl}" alt="Image" class="w-full max-h-[400px] object-cover" loading="lazy" onerror="this.style.display='none'"></div>`;
            }
            content = payload.caption || payload.text || payload.content || '';
        } else {
            content = payload.content || payload.text || payload.body || '';
            if (payload.title) {
                content = `<span class="font-semibold text-white">${this.escapeHtml(payload.title)}</span><br><span class="text-gray-300">${content}</span>`;
            }
        }

        if (payload.url && !kind.startsWith('video') && !kind.startsWith('image')) {
            content += `<div class="mt-2"><a href="${this.escapeHtml(payload.url)}" target="_blank" rel="noopener" class="text-cyan-400 hover:underline text-xs truncate block">🔗 ${this.escapeHtml(payload.url)}</a></div>`;
        }

        return `
        <article class="node-card bg-white/[0.03] border border-white/[0.06] rounded-xl p-4 hover:border-white/10 hover:bg-white/[0.05] transition-all" data-kind="${kind}">
            <div class="flex items-center gap-3 mb-3">
                <a href="profile.html?pubkey=${node.author_pubkey}" class="w-10 h-10 rounded-full flex items-center justify-center text-xs font-bold text-white flex-shrink-0 hover:ring-2 ring-white/20 transition" style="background: hsl(${hue}, 55%, 45%)">
                    ${authorShort.substring(0, 2).toUpperCase()}
                </a>
                <div class="flex-1 min-w-0">
                    <a href="profile.html?pubkey=${node.author_pubkey}" class="text-sm font-semibold text-white hover:underline">${authorShort}…</a>
                    <span class="text-xs text-gray-500 ml-2">${time}</span>
                </div>
                <span class="px-2 py-1 rounded-md text-[11px] font-medium ${badgeColor}">${badgeIcon} ${badgeLabel}</span>
            </div>
            ${mediaSection}
            <div class="text-sm text-gray-300 leading-relaxed ${content ? '' : 'text-gray-600 italic'}">${content || 'Empty payload'}</div>
            <div class="flex items-center gap-4 mt-3 pt-3 border-t border-white/5 text-xs">
                <a href="thread.html?id=${node.id}" class="text-gray-500 hover:text-cyan-400 transition flex items-center gap-1">💬 Thread</a>
                <span class="text-gray-600 font-mono ml-auto" title="${node.id}">ID: ${(node.id || '').substring(0, 8)}</span>
            </div>
        </article>`;
    },

    getKindInfo(kind) {
        const map = {
            'text': ['📝', 'Text', 'bg-blue-500/15 text-blue-400'],
            'text/article': ['📰', 'Article', 'bg-purple-500/15 text-purple-400'],
            'article': ['📰', 'Article', 'bg-purple-500/15 text-purple-400'],
            'note': ['📝', 'Note', 'bg-blue-500/15 text-blue-400'],
            'comment': ['💬', 'Comment', 'bg-gray-500/15 text-gray-400'],
            'video': ['📹', 'Video', 'bg-red-500/15 text-red-400'],
            'video/external': ['📺', 'Video', 'bg-red-500/15 text-red-400'],
            'video/embed': ['📺', 'Embed', 'bg-red-500/15 text-red-400'],
            'image': ['📸', 'Image', 'bg-pink-500/15 text-pink-400'],
            'image/url': ['📸', 'Image', 'bg-pink-500/15 text-pink-400'],
            'photo': ['📸', 'Photo', 'bg-pink-500/15 text-pink-400'],
            'link': ['🔗', 'Link', 'bg-yellow-500/15 text-yellow-400'],
            'bookmark': ['🔖', 'Bookmark', 'bg-yellow-500/15 text-yellow-400'],
        };
        return map[kind] || ['📦', kind, 'bg-gray-500/15 text-gray-400'];
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

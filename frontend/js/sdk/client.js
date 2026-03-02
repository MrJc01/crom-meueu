/**
 * Crom Protocol Client SDK
 * Handles API interactions.
 */
class CromClient {
    constructor(baseUrl = '') {
        this.baseUrl = baseUrl;
    }

    setBaseUrl(url) {
        this.baseUrl = url;
    }

    /**
     * Query Nodes from the Network
     * @param {Object} filters - { author, kind, tags: [], limit }
     * @returns {Promise<Array>} List of nodes
     */
    async query(filters = {}) {
        const url = `${this.baseUrl}/v1/query`;
        try {
            const response = await fetch(url, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    filters: filters,
                    limit: filters.limit || 50
                })
            });

            if (!response.ok) throw new Error(`API Error: ${response.status}`);
            return await response.json();
        } catch (error) {
            console.error('CromClient Query Error:', error);
            throw error;
        }
    }

    /**
     * Publish a new Node
     * @param {Object} nodeData - { author_pubkey, kind, payload, signature, claimed_at }
     * @returns {Promise<Object>} { id, status }
     */
    async publish(nodeData) {
        const url = `${this.baseUrl}/v1/publish`;
        try {
            const response = await fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-MeuEu-Author': nodeData.author_pubkey || ''
                },
                body: JSON.stringify(nodeData)
            });

            if (!response.ok) {
                const errText = await response.text();
                throw new Error(`Publish Error: ${response.status} - ${errText}`);
            }
            return await response.json();
        } catch (error) {
            console.error('CromClient Publish Error:', error);
            throw error;
        }
    }

    /**
     * Import a batch of Historical Nodes (Backup Restore)
     * @param {Array} nodes - Array of previously signed nodeData
     * @param {String} authorPubKey - The pubkey claiming the restoration
     */
    async importBulk(nodes, authorPubKey) {
        const url = `${this.baseUrl}/v1/import`;
        try {
            const response = await fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-MeuEu-Author': authorPubKey || ''
                },
                body: JSON.stringify(nodes)
            });

            if (!response.ok) {
                const errText = await response.text();
                throw new Error(`Import Error: ${response.status} - ${errText}`);
            }
            return await response.json();
        } catch (error) {
            console.error('CromClient Import Error:', error);
            throw error;
        }
    }
}

// Export global
window.CromClient = CromClient;

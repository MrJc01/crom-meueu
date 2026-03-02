/**
 * Crom Protocol Universal Backup & Restore
 * Dependencies: jszip.min.js, sdk/auth.js, sdk/client.js
 */

class CromExporter {
    constructor() { }

    /**
     * Creates a ZIP file containing the user's .cromid and all their published nodes.
     */
    async exportBackup() {
        if (!window.cromAuth || !window.cromAuth.pubKeyHex) {
            alert("Please login first to export your data.");
            return;
        }

        const pubKey = window.cromAuth.pubKeyHex;
        const cromidData = window.cromAuth._cromidData || JSON.parse(sessionStorage.getItem('crom_vault') || '{}');

        if (!cromidData.pubKey) {
            alert("Cannot find valid vault data in session.");
            return;
        }

        // UI Feedback
        const btn = document.getElementById('export-backup-btn');
        const ogText = btn ? btn.innerText : "";
        if (btn) btn.innerText = "⏳ Gathering History...";

        try {
            // Fetch everything the user has published (Iterative if pagination exists, for now max 1000)
            const history = await window.cromClient.query({ author: pubKey, limit: 1000 });

            if (btn) btn.innerText = "📦 Compressing Data...";

            // Initialize JSZip
            const zip = new JSZip();

            // Add Vault Identity
            zip.file("identity.cromid", JSON.stringify(cromidData, null, 2));

            // Add History
            zip.file("history.json", JSON.stringify(history, null, 2));

            // Generate ZIP
            const blob = await zip.generateAsync({ type: "blob" });

            // Trigger Download
            const dateStr = new Date().toISOString().split('T')[0];
            const url = URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.href = url;
            a.download = `CromBackup_${pubKey.substring(0, 6)}_${dateStr}.zip`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);

            if (btn) btn.innerText = "✅ Export Successful!";
            setTimeout(() => { if (btn) btn.innerText = ogText; }, 3000);

        } catch (err) {
            console.error(err);
            alert("Failed to export backup: " + err.message);
            if (btn) btn.innerText = "❌ Export Failed";
            setTimeout(() => { if (btn) btn.innerText = ogText; }, 3000);
        }
    }

    /**
     * Reads a provided ZIP file, extracts identity, loads it, and republicates history.
     */
    async restoreBackup(file) {
        if (!file) return;

        // UI Feedback
        const label = document.getElementById('restore-file-label');
        if (label) label.innerHTML = "⏳ Scanning Backup...";

        try {
            const zip = await JSZip.loadAsync(file);

            // 1. Locate and unlock Vault
            const vaultFile = zip.file("identity.cromid");
            if (!vaultFile) throw new Error("identity.cromid not found in ZIP backup.");

            const vaultDataStr = await vaultFile.async("string");
            const cromidData = JSON.parse(vaultDataStr);

            // Log user in utilizing Auth SDK rules
            await window.cromAuth.loadIdentityAuto(cromidData);
            sessionStorage.setItem('crom_vault', JSON.stringify(cromidData));

            // Clean interface login
            if (typeof updateAuthUI === 'function') updateAuthUI();

            // 2. Read History
            if (label) label.innerHTML = "⏳ Importing Nodes...";
            const historyFile = zip.file("history.json");

            if (historyFile) {
                const historyDataStr = await historyFile.async("string");
                const historyNodes = JSON.parse(historyDataStr);

                // Republicate in chunks of 100 to avoid Payload Too Large errors and bypass 1-by-1 rate limits
                const chunkSize = 100;
                let processed = 0;

                for (let i = 0; i < historyNodes.length; i += chunkSize) {
                    const chunk = historyNodes.slice(i, i + chunkSize);
                    try {
                        const res = await window.cromClient.importBulk(chunk, window.cromAuth.pubKeyHex);
                        processed += res.imported || 0;
                        if (label) label.innerHTML = `⏳ Importing (${processed}/${historyNodes.length})...`;
                    } catch (e) {
                        console.warn("Failed to import chunk:", e);
                    }
                }
                alert(`Backup restored successfully! Synced ${processed} historical nodes.`);
            } else {
                alert("Vault loaded, but no history.json found inside backup.");
            }

            if (label) label.innerHTML = "✅ Restored";
            window.location.reload();

        } catch (err) {
            console.error(err);
            alert("Failed to restore backup: " + err.message);
            if (label) label.innerHTML = "❌ Restore Failed";
            setTimeout(() => { if (label) label.innerHTML = "📂 Import Backup (.zip)"; }, 3000);
        }
    }
}

window.cromExporter = new CromExporter();

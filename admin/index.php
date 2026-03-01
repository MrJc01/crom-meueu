<?php
/**
 * Crom-Meueu Admin Dashboard v2
 * Login with Admin Token + .cromid file upload
 */
session_start();
$defaultServer = 'http://localhost:8080';

function apiCall($method, $endpoint, $data = null)
{
    $server = $_SESSION['server_url'] ?? $GLOBALS['defaultServer'];
    $token = $_SESSION['admin_token'] ?? '';
    $url = rtrim($server, '/') . $endpoint;
    $opts = [
        'http' => [
            'method' => strtoupper($method),
            'header' => "Content-Type: application/json\r\nX-Admin-Token: {$token}\r\n",
            'timeout' => 10,
            'ignore_errors' => true,
        ]
    ];
    if ($data !== null && in_array(strtoupper($method), ['POST', 'PUT'])) {
        $opts['http']['content'] = json_encode($data);
    }
    $ctx = stream_context_create($opts);
    $response = @file_get_contents($url, false, $ctx);
    $statusCode = 0;
    if (isset($http_response_header[0])) {
        preg_match('/(\d{3})/', $http_response_header[0], $m);
        $statusCode = intval($m[1] ?? 0);
    }
    return [
        'status' => $statusCode,
        'body' => $response !== false ? $response : '',
        'json' => $response ? json_decode($response, true) : null,
    ];
}

// ─── Actions ─────────────────────────────────────────────────────
$flash = '';
$flashType = 'success';
$action = $_POST['action'] ?? $_GET['action'] ?? '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    switch ($action) {
        case 'login':
            $_SESSION['admin_token'] = $_POST['token'] ?? '';
            $_SESSION['server_url'] = $_POST['server_url'] ?? $defaultServer;
            $hasCromId = false;
            // Store .cromid data if uploaded
            if (!empty($_POST['cromid_json'])) {
                $cromidData = json_decode($_POST['cromid_json'], true);
                if ($cromidData && isset($cromidData['pubKey'])) {
                    $_SESSION['cromid'] = $cromidData;
                    $_SESSION['pubkey'] = $cromidData['pubKey'];
                    $_SESSION['trusted_source'] = $cromidData['trusted_source'] ?? '';
                    $hasCromId = true;
                }
            }
            // Auth: .cromid alone is enough to enter. Token enables admin API calls.
            if ($hasCromId && empty($_POST['token'])) {
                $_SESSION['authenticated'] = true;
                $flash = '✅ Identidade carregada! (Modo somente leitura — insira o token para operações admin)';
            } elseif (!empty($_POST['token'])) {
                $test = apiCall('GET', '/admin/stats');
                if ($test['status'] === 200) {
                    $_SESSION['authenticated'] = true;
                    $flash = '✅ Conectado com acesso total!';
                } else {
                    // Token failed but .cromid present — allow identity-only access
                    if ($hasCromId) {
                        $_SESSION['authenticated'] = true;
                        $flash = '⚠️ Token inválido, mas identidade carregada. Modo leitura.';
                        $flashType = 'error';
                    } else {
                        $_SESSION['authenticated'] = false;
                        unset($_SESSION['admin_token'], $_SESSION['cromid'], $_SESSION['pubkey']);
                        $flash = '❌ Token inválido e nenhum .cromid fornecido.';
                        $flashType = 'error';
                    }
                }
            } else {
                $flash = '❌ Forneça um ficheiro .cromid ou um token admin.';
                $flashType = 'error';
            }
            break;
        case 'logout':
            session_destroy();
            header('Location: index.php');
            exit;
        case 'change_server':
            $_SESSION['server_url'] = $_POST['server_url'] ?? $defaultServer;
            $test = apiCall('GET', '/admin/stats');
            $flash = $test['status'] === 200 ? '✅ Servidor alterado!' : '⚠️ Conexão falhou.';
            $flashType = $test['status'] === 200 ? 'success' : 'error';
            break;
        case 'add_whitelist':
            $r = apiCall('POST', '/admin/whitelist', ['public_key' => $_POST['public_key']]);
            $flash = $r['status'] === 200 ? '✅ Adicionado.' : '❌ ' . $r['body'];
            $flashType = $r['status'] === 200 ? 'success' : 'error';
            break;
        case 'remove_whitelist':
            $r = apiCall('DELETE', '/admin/whitelist?public_key=' . urlencode($_POST['public_key']));
            $flash = $r['status'] === 200 ? '✅ Removido.' : '❌ ' . $r['body'];
            $flashType = $r['status'] === 200 ? 'success' : 'error';
            break;
        case 'ban_word':
            $r = apiCall('POST', '/admin/banned_words', ['word' => $_POST['word']]);
            $flash = $r['status'] === 200 ? '✅ Banido.' : '❌ ' . $r['body'];
            $flashType = $r['status'] === 200 ? 'success' : 'error';
            break;
        case 'unban_word':
            $r = apiCall('DELETE', '/admin/banned_words?word=' . urlencode($_POST['word']));
            $flash = $r['status'] === 200 ? '✅ Removido.' : '❌ ' . $r['body'];
            $flashType = $r['status'] === 200 ? 'success' : 'error';
            break;
        case 'ban_user':
            $r = apiCall('POST', '/admin/banned_users', ['public_key' => $_POST['public_key'], 'reason' => $_POST['reason'] ?? '']);
            $flash = $r['status'] === 200 ? '✅ Banido.' : '❌ ' . $r['body'];
            $flashType = $r['status'] === 200 ? 'success' : 'error';
            break;
        case 'unban_user':
            $r = apiCall('DELETE', '/admin/banned_users?public_key=' . urlencode($_POST['public_key']));
            $flash = $r['status'] === 200 ? '✅ Desbanido.' : '❌ ' . $r['body'];
            $flashType = $r['status'] === 200 ? 'success' : 'error';
            break;
        case 'ban_hash':
            $r = apiCall('POST', '/admin/banned_hashes', ['hash' => $_POST['hash'], 'reason' => $_POST['reason'] ?? '']);
            $flash = $r['status'] === 200 ? '✅ Hash banido.' : '❌ ' . $r['body'];
            $flashType = $r['status'] === 200 ? 'success' : 'error';
            break;
        case 'unban_hash':
            $r = apiCall('DELETE', '/admin/banned_hashes?hash=' . urlencode($_POST['hash']));
            $flash = $r['status'] === 200 ? '✅ Removido.' : '❌ ' . $r['body'];
            $flashType = $r['status'] === 200 ? 'success' : 'error';
            break;
        case 'delete_node':
            $r = apiCall('DELETE', '/admin/node?id=' . urlencode($_POST['node_id']));
            $flash = $r['status'] === 200 ? '✅ Eliminado.' : '❌ ' . $r['body'];
            $flashType = $r['status'] === 200 ? 'success' : 'error';
            break;
    }
}

$authenticated = $_SESSION['authenticated'] ?? false;
$serverUrl = $_SESSION['server_url'] ?? $defaultServer;
$page = $_GET['page'] ?? 'dashboard';
$sessionPubKey = $_SESSION['pubkey'] ?? '';

$stats = $whitelist = $bannedWords = $bannedUsers = $bannedHashes = $peers = $meta = null;
if ($authenticated) {
    switch ($page) {
        case 'dashboard':
            $stats = apiCall('GET', '/admin/stats')['json'];
            $meta = apiCall('GET', '/meta')['json'];
            break;
        case 'whitelist':
            $whitelist = apiCall('GET', '/admin/whitelist')['json'] ?? [];
            break;
        case 'banned_words':
            $bannedWords = apiCall('GET', '/admin/banned_words')['json'] ?? [];
            break;
        case 'banned_users':
            $bannedUsers = apiCall('GET', '/admin/banned_users')['json'] ?? [];
            break;
        case 'banned_hashes':
            $bannedHashes = apiCall('GET', '/admin/banned_hashes')['json'] ?? [];
            break;
        case 'peers':
            $peers = apiCall('GET', '/v1/peers')['json'] ?? [];
            break;
    }
}
?>
<!DOCTYPE html>
<html lang="pt">

<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Crom Admin</title>
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap');

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box
        }

        :root {
            --bg: #0a0b0f;
            --card: #12141c;
            --input: #1a1d28;
            --border: #2a2d3a;
            --text: #e0e0e0;
            --muted: #8a8d9a;
            --accent: #00d2ff;
            --accent-d: #0094b3;
            --danger: #ff4444;
            --success: #00ff88;
            --warn: #ffaa00;
            --sidebar: 260px;
        }

        body {
            font-family: 'Inter', sans-serif;
            background: var(--bg);
            color: var(--text);
            min-height: 100vh
        }

        /* Login */
        .login-wrap {
            display: flex;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
            background: radial-gradient(ellipse at top, #111428, #0a0b0f 60%)
        }

        .login-box {
            background: var(--card);
            border: 1px solid var(--border);
            border-radius: 16px;
            padding: 40px;
            width: 100%;
            max-width: 480px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, .5)
        }

        .login-box h1 {
            font-size: 1.8em;
            margin-bottom: 6px;
            background: linear-gradient(135deg, #fff, var(--accent));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent
        }

        .login-box .sub {
            color: var(--muted);
            font-size: .9em;
            margin-bottom: 25px
        }

        /* Layout */
        .layout {
            display: flex;
            min-height: 100vh
        }

        .sidebar {
            width: var(--sidebar);
            background: var(--card);
            border-right: 1px solid var(--border);
            padding: 20px 0;
            position: fixed;
            height: 100vh;
            overflow-y: auto;
            display: flex;
            flex-direction: column
        }

        .sb-head {
            padding: 0 20px 15px;
            border-bottom: 1px solid var(--border);
            margin-bottom: 10px
        }

        .sb-head h2 {
            font-size: 1.1em;
            background: linear-gradient(135deg, #fff, var(--accent));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent
        }

        .sb-head small {
            color: var(--muted);
            font-size: .7em;
            font-family: 'JetBrains Mono', monospace
        }

        .sb-id {
            margin: 8px 20px 0;
            padding: 8px 12px;
            background: var(--input);
            border-radius: 8px;
            font-size: .7em
        }

        .sb-id .lbl {
            color: var(--muted);
            text-transform: uppercase;
            letter-spacing: 1px;
            font-size: .65em
        }

        .sb-id .val {
            color: var(--accent);
            font-family: 'JetBrains Mono', monospace;
            word-break: break-all;
            margin-top: 2px
        }

        .nav-sec {
            padding: 10px 0
        }

        .nav-sec h4 {
            padding: 5px 20px;
            font-size: .7em;
            text-transform: uppercase;
            letter-spacing: 1.5px;
            color: var(--muted);
            margin-bottom: 5px
        }

        .nav-a {
            display: flex;
            align-items: center;
            gap: 10px;
            padding: 10px 20px;
            color: var(--muted);
            text-decoration: none;
            font-size: .9em;
            transition: .2s;
            border-left: 3px solid transparent
        }

        .nav-a:hover {
            background: rgba(0, 210, 255, .05);
            color: var(--text)
        }

        .nav-a.on {
            background: rgba(0, 210, 255, .08);
            color: var(--accent);
            border-left-color: var(--accent)
        }

        .nav-a .ic {
            font-size: 1.1em;
            width: 24px;
            text-align: center
        }

        .sb-foot {
            margin-top: auto;
            padding: 15px 20px;
            border-top: 1px solid var(--border)
        }

        .main {
            margin-left: var(--sidebar);
            padding: 30px;
            width: calc(100% - var(--sidebar))
        }

        /* Components */
        .ph {
            margin-bottom: 25px
        }

        .ph h1 {
            font-size: 1.6em;
            font-weight: 600
        }

        .ph p {
            color: var(--muted);
            font-size: .9em;
            margin-top: 4px
        }

        .card {
            background: var(--card);
            border: 1px solid var(--border);
            border-radius: 12px;
            padding: 24px;
            margin-bottom: 20px
        }

        .card h3 {
            font-size: 1em;
            font-weight: 600;
            margin-bottom: 15px
        }

        .sg {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 16px
        }

        .sc {
            background: var(--input);
            border: 1px solid var(--border);
            border-radius: 10px;
            padding: 20px;
            text-align: center
        }

        .sc .v {
            font-size: 2.2em;
            font-weight: 700;
            color: var(--accent);
            font-family: 'JetBrains Mono', monospace
        }

        .sc .l {
            font-size: .8em;
            color: var(--muted);
            margin-top: 5px;
            text-transform: uppercase;
            letter-spacing: 1px
        }

        .mg {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
            gap: 12px;
            margin-top: 15px
        }

        .mi {
            background: var(--input);
            border-radius: 8px;
            padding: 12px 16px
        }

        .mi .k {
            font-size: .7em;
            text-transform: uppercase;
            letter-spacing: 1px;
            color: var(--muted)
        }

        .mi .vl {
            font-family: 'JetBrains Mono', monospace;
            font-size: .85em;
            color: var(--accent);
            margin-top: 4px;
            word-break: break-all
        }

        /* Forms */
        .fg {
            margin-bottom: 15px
        }

        .fg label {
            display: block;
            font-size: .8em;
            color: var(--muted);
            margin-bottom: 6px;
            text-transform: uppercase;
            letter-spacing: .5px
        }

        input[type="text"],
        input[type="password"],
        input[type="url"],
        select {
            width: 100%;
            padding: 12px 14px;
            background: var(--input);
            border: 1px solid var(--border);
            border-radius: 8px;
            color: var(--text);
            font-family: 'JetBrains Mono', monospace;
            font-size: .9em;
            transition: .2s
        }

        input:focus,
        select:focus {
            outline: none;
            border-color: var(--accent);
            box-shadow: 0 0 0 3px rgba(0, 210, 255, .1)
        }

        .inf {
            display: flex;
            gap: 10px;
            align-items: flex-end
        }

        .inf .fg {
            flex: 1;
            margin-bottom: 0
        }

        .btn {
            padding: 10px 20px;
            border: none;
            border-radius: 8px;
            font-size: .85em;
            font-weight: 600;
            cursor: pointer;
            transition: .2s;
            font-family: 'Inter', sans-serif
        }

        .btn-p {
            background: var(--accent);
            color: #000
        }

        .btn-p:hover {
            background: var(--accent-d);
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(0, 210, 255, .3)
        }

        .btn-d {
            background: var(--danger);
            color: #fff
        }

        .btn-d:hover {
            background: #c22
        }

        .btn-s {
            padding: 6px 14px;
            font-size: .75em
        }

        .btn-g {
            background: transparent;
            border: 1px solid var(--border);
            color: var(--muted)
        }

        .btn-g:hover {
            border-color: var(--danger);
            color: var(--danger)
        }

        .btn-out {
            width: 100%;
            background: transparent;
            border: 1px solid var(--border);
            color: var(--danger);
            padding: 8px;
            border-radius: 8px;
            cursor: pointer;
            font-size: .85em;
            transition: .2s
        }

        .btn-out:hover {
            background: rgba(255, 68, 68, .1);
            border-color: var(--danger)
        }

        /* Tables */
        .dt {
            width: 100%;
            border-collapse: collapse;
            font-size: .85em
        }

        .dt th {
            text-align: left;
            padding: 10px 12px;
            font-size: .75em;
            text-transform: uppercase;
            letter-spacing: 1px;
            color: var(--muted);
            border-bottom: 1px solid var(--border)
        }

        .dt td {
            padding: 10px 12px;
            border-bottom: 1px solid rgba(42, 45, 58, .5);
            font-family: 'JetBrains Mono', monospace;
            font-size: .9em;
            word-break: break-all
        }

        .dt tr:hover td {
            background: rgba(0, 210, 255, .03)
        }

        .badge {
            display: inline-block;
            padding: 3px 10px;
            border-radius: 20px;
            font-size: .75em;
            font-weight: 600
        }

        .bg {
            background: rgba(0, 255, 136, .15);
            color: var(--success)
        }

        .bw {
            background: rgba(255, 170, 0, .15);
            color: var(--warn)
        }

        .bb {
            background: rgba(255, 68, 68, .15);
            color: var(--danger)
        }

        .empty {
            text-align: center;
            padding: 40px;
            color: var(--muted)
        }

        .empty .ic {
            font-size: 2em;
            margin-bottom: 10px
        }

        /* Flash */
        .flash {
            padding: 12px 16px;
            border-radius: 8px;
            margin-bottom: 20px;
            font-size: .9em;
            border: 1px solid
        }

        .flash-success {
            background: rgba(0, 255, 136, .08);
            border-color: rgba(0, 255, 136, .2);
            color: var(--success)
        }

        .flash-error {
            background: rgba(255, 68, 68, .08);
            border-color: rgba(255, 68, 68, .2);
            color: var(--danger)
        }

        /* Dropzone */
        .dz {
            border: 2px dashed var(--border);
            border-radius: 12px;
            padding: 30px;
            text-align: center;
            cursor: pointer;
            transition: .3s;
            background: var(--input)
        }

        .dz:hover,
        .dz.over {
            border-color: var(--accent);
            background: rgba(0, 210, 255, .05)
        }

        .dz .ic {
            font-size: 2.5em;
            margin-bottom: 8px
        }

        .dz p {
            color: var(--muted);
            font-size: .85em
        }

        .dz-ok {
            background: rgba(0, 255, 136, .05);
            border-color: var(--success)
        }

        .dz-ok .ic {
            color: var(--success)
        }

        .dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            background: var(--success);
            display: inline-block;
            animation: pulse 2s infinite
        }

        @keyframes pulse {

            0%,
            100% {
                opacity: 1
            }

            50% {
                opacity: .4
            }
        }

        @media(max-width:768px) {
            .sidebar {
                display: none
            }

            .main {
                margin-left: 0;
                width: 100%;
                padding: 15px
            }

            .sg {
                grid-template-columns: 1fr
            }

            .inf {
                flex-direction: column
            }
        }
    </style>
</head>

<body>

    <?php if (!$authenticated): ?>
        <!-- ═══════════ LOGIN ═══════════ -->
        <div class="login-wrap">
            <div class="login-box" style="text-align:center">
                <div style="font-size:3em;margin-bottom:10px">🛡️</div>
                <h1>Crom Admin</h1>
                <p class="sub">Painel de Controle do Nó Meueu</p>

                <?php if ($flash): ?>
                    <div class="flash flash-<?= $flashType ?>"><?= htmlspecialchars($flash) ?></div>
                <?php endif; ?>

                <form method="POST" id="login-form">
                    <input type="hidden" name="action" value="login">
                    <input type="hidden" name="cromid_json" id="cromid_json" value="">

                    <div class="fg">
                        <label>Servidor Backend</label>
                        <input type="url" name="server_url" value="<?= htmlspecialchars($serverUrl) ?>"
                            placeholder="http://localhost:8080">
                    </div>

                    <div class="fg">
                        <label>Token de Admin <span
                                style="color:var(--muted);text-transform:none;letter-spacing:0">(opcional se tiver
                                .cromid)</span></label>
                        <input type="password" name="token" placeholder="Senha do administrador">
                    </div>

                    <div class="fg">
                        <label>Ficheiro .cromid (Identidade)</label>
                        <div class="dz" id="login-dropzone" onclick="document.getElementById('login-cromid-file').click()">
                            <div class="ic" id="dz-icon">📂</div>
                            <p id="dz-text">Arraste o .cromid aqui ou clique para selecionar</p>
                            <input type="file" id="login-cromid-file" accept=".cromid,.json" style="display:none"
                                onchange="handleCromIdFile(this)">
                        </div>
                    </div>

                    <button type="submit" class="btn btn-p" style="width:100%;margin-top:10px;padding:14px">
                        🔐 Conectar
                    </button>
                </form>
            </div>
        </div>

        <script>
            const dz = document.getElementById('login-dropzone');
            dz.addEventListener('dragover', e => { e.preventDefault(); dz.classList.add('over'); });
            dz.addEventListener('dragleave', () => dz.classList.remove('over'));
            dz.addEventListener('drop', e => {
                e.preventDefault(); dz.classList.remove('over');
                if (e.dataTransfer.files.length > 0) parseCromId(e.dataTransfer.files[0]);
            });

            function handleCromIdFile(input) {
                if (input.files.length > 0) parseCromId(input.files[0]);
            }

            function parseCromId(file) {
                const reader = new FileReader();
                reader.onload = function (e) {
                    try {
                        const data = JSON.parse(e.target.result);
                        if (!data.pubKey) throw new Error('Campo pubKey não encontrado');
                        document.getElementById('cromid_json').value = e.target.result;
                        document.getElementById('dz-icon').textContent = '✅';
                        document.getElementById('dz-text').innerHTML =
                            '<b style="color:var(--success)">Identidade carregada!</b><br>' +
                            '<span style="font-family:JetBrains Mono,monospace;font-size:.75em;color:var(--accent)">' +
                            data.pubKey.substring(0, 24) + '...</span>';
                        dz.classList.add('dz-ok');
                    } catch (err) {
                        alert('❌ Ficheiro inválido: ' + err.message);
                    }
                };
                reader.readAsText(file);
            }
        </script>

    <?php else: ?>
        <!-- ═══════════ DASHBOARD ═══════════ -->
        <div class="layout">
            <nav class="sidebar">
                <div class="sb-head">
                    <h2>🛡️ Crom Admin</h2>
                    <small><?= htmlspecialchars(parse_url($serverUrl, PHP_URL_HOST) ?: $serverUrl) ?></small>
                </div>

                <?php if ($sessionPubKey): ?>
                    <div class="sb-id">
                        <div class="lbl">🔑 Identidade</div>
                        <div class="val"><?= htmlspecialchars(substr($sessionPubKey, 0, 20)) ?>...</div>
                    </div>
                <?php endif; ?>

                <div class="nav-sec">
                    <h4>Principal</h4>
                    <a href="?page=dashboard" class="nav-a <?= $page === 'dashboard' ? 'on' : '' ?>"><span
                            class="ic">📊</span> Dashboard</a>
                    <a href="?page=peers" class="nav-a <?= $page === 'peers' ? 'on' : '' ?>"><span class="ic">🌐</span>
                        Peers</a>
                </div>
                <div class="nav-sec">
                    <h4>Controle de Acesso</h4>
                    <a href="?page=whitelist" class="nav-a <?= $page === 'whitelist' ? 'on' : '' ?>"><span
                            class="ic">✅</span> Whitelist</a>
                    <a href="?page=banned_users" class="nav-a <?= $page === 'banned_users' ? 'on' : '' ?>"><span
                            class="ic">⛔</span> Banidos</a>
                </div>
                <div class="nav-sec">
                    <h4>Moderação</h4>
                    <a href="?page=banned_words" class="nav-a <?= $page === 'banned_words' ? 'on' : '' ?>"><span
                            class="ic">🚫</span> Palavras</a>
                    <a href="?page=banned_hashes" class="nav-a <?= $page === 'banned_hashes' ? 'on' : '' ?>"><span
                            class="ic">🔒</span> Hashes</a>
                    <a href="?page=delete" class="nav-a <?= $page === 'delete' ? 'on' : '' ?>"><span class="ic">🗑️</span>
                        Eliminar</a>
                </div>
                <div class="nav-sec">
                    <h4>Identidade</h4>
                    <a href="?page=identity" class="nav-a <?= $page === 'identity' ? 'on' : '' ?>"><span
                            class="ic">🔑</span> Cofre .cromid</a>
                </div>
                <div class="nav-sec">
                    <h4>Config</h4>
                    <a href="?page=server" class="nav-a <?= $page === 'server' ? 'on' : '' ?>"><span class="ic">⚙️</span>
                        Servidor</a>
                </div>
                <div class="sb-foot">
                    <form method="POST"><input type="hidden" name="action" value="logout">
                        <button type="submit" class="btn-out">🔓 Sair</button>
                    </form>
                </div>
            </nav>

            <main class="main">
                <?php if ($flash): ?>
                    <div class="flash flash-<?= $flashType ?>"><?= htmlspecialchars($flash) ?></div>
                <?php endif; ?>

                <?php if ($page === 'dashboard'): ?>
                    <div class="ph">
                        <h1>📊 Dashboard</h1>
                        <p>Visão geral do nó</p>
                    </div>
                    <?php if ($meta): ?>
                        <div class="card">
                            <h3>Metadados do Nó</h3>
                            <div class="mg">
                                <?php foreach ($meta as $k => $v): ?>
                                    <div class="mi">
                                        <div class="k"><?= htmlspecialchars($k) ?></div>
                                        <div class="vl"><?= htmlspecialchars($v ?: '—') ?></div>
                                    </div>
                                <?php endforeach; ?>
                            </div>
                        </div>
                    <?php endif; ?>
                    <div class="sg">
                        <div class="sc">
                            <div class="v"><?= number_format($stats['total_nodes'] ?? 0) ?></div>
                            <div class="l">Nodes</div>
                        </div>
                        <div class="sc">
                            <div class="v"><?= number_format($stats['total_users'] ?? 0) ?></div>
                            <div class="l">Users</div>
                        </div>
                        <div class="sc">
                            <div class="v"><?= number_format($stats['active_nodes_24h'] ?? 0) ?></div>
                            <div class="l">Ativos 24h</div>
                        </div>
                    </div>

                <?php elseif ($page === 'whitelist'): ?>
                    <div class="ph">
                        <h1>✅ Whitelist</h1>
                        <p>Chaves autorizadas (modo WHITELIST)</p>
                    </div>
                    <div class="card">
                        <h3>Adicionar</h3>
                        <form method="POST" class="inf">
                            <input type="hidden" name="action" value="add_whitelist">
                            <div class="fg"><label>Chave Pública</label>
                                <input type="text" name="public_key" placeholder="abc123..." maxlength="64" required
                                    value="<?= htmlspecialchars($_GET['prefill'] ?? $sessionPubKey) ?>">
                            </div>
                            <button type="submit" class="btn btn-p">Adicionar</button>
                        </form>
                    </div>
                    <div class="card">
                        <h3>Lista (<?= count($whitelist) ?>)</h3>
                        <?php if (empty($whitelist)): ?>
                            <div class="empty">
                                <div class="ic">📭</div>
                                <p>Vazia</p>
                            </div>
                        <?php else: ?>
                            <table class="dt">
                                <thead>
                                    <tr>
                                        <th>Chave Pública</th>
                                        <th>Ação</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <?php foreach ($whitelist as $k): ?>
                                        <tr>
                                            <td><?= htmlspecialchars($k) ?></td>
                                            <td>
                                                <form method="POST" style="display:inline" onsubmit="return confirm('Remover?')">
                                                    <input type="hidden" name="action" value="remove_whitelist">
                                                    <input type="hidden" name="public_key" value="<?= htmlspecialchars($k) ?>">
                                                    <button type="submit" class="btn btn-g btn-s">Remover</button>
                                                </form>
                                            </td>
                                        </tr>
                                    <?php endforeach; ?>
                                </tbody>
                            </table>
                        <?php endif; ?>
                    </div>

                <?php elseif ($page === 'banned_words'): ?>
                    <div class="ph">
                        <h1>🚫 Palavras Banidas</h1>
                        <p>Filtro de conteúdo</p>
                    </div>
                    <div class="card">
                        <h3>Banir Palavra</h3>
                        <form method="POST" class="inf">
                            <input type="hidden" name="action" value="ban_word">
                            <div class="fg"><label>Palavra</label><input type="text" name="word" placeholder="spam" required>
                            </div>
                            <button type="submit" class="btn btn-d">Banir</button>
                        </form>
                    </div>
                    <div class="card">
                        <h3>Atuais (<?= count($bannedWords) ?>)</h3>
                        <?php if (empty($bannedWords)): ?>
                            <div class="empty">
                                <div class="ic">✨</div>
                                <p>Nenhuma</p>
                            </div>
                        <?php else: ?>
                            <table class="dt">
                                <thead>
                                    <tr>
                                        <th>Palavra</th>
                                        <th>Ação</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <?php foreach ($bannedWords as $w): ?>
                                        <tr>
                                            <td><?= htmlspecialchars($w) ?></td>
                                            <td>
                                                <form method="POST" style="display:inline" onsubmit="return confirm('Remover?')">
                                                    <input type="hidden" name="action" value="unban_word">
                                                    <input type="hidden" name="word" value="<?= htmlspecialchars($w) ?>">
                                                    <button type="submit" class="btn btn-g btn-s">Remover</button>
                                                </form>
                                            </td>
                                        </tr>
                                    <?php endforeach; ?>
                                </tbody>
                            </table>
                        <?php endif; ?>
                    </div>

                <?php elseif ($page === 'banned_users'): ?>
                    <div class="ph">
                        <h1>⛔ Utilizadores Banidos</h1>
                        <p>Chaves bloqueadas</p>
                    </div>
                    <div class="card">
                        <h3>Banir</h3>
                        <form method="POST">
                            <input type="hidden" name="action" value="ban_user">
                            <div class="fg"><label>Chave Pública</label>
                                <input type="text" name="public_key" placeholder="abc123..." maxlength="64" required
                                    value="<?= htmlspecialchars($_GET['prefill'] ?? '') ?>">
                            </div>
                            <div class="fg"><label>Motivo</label><input type="text" name="reason" placeholder="Spam"></div>
                            <button type="submit" class="btn btn-d">Banir</button>
                        </form>
                    </div>
                    <div class="card">
                        <h3>Banidos (<?= count($bannedUsers) ?>)</h3>
                        <?php if (empty($bannedUsers)): ?>
                            <div class="empty">
                                <div class="ic">👤</div>
                                <p>Nenhum</p>
                            </div>
                        <?php else: ?>
                            <table class="dt">
                                <thead>
                                    <tr>
                                        <th>Chave Pública</th>
                                        <th>Ação</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <?php foreach ($bannedUsers as $u): ?>
                                        <tr>
                                            <td><?= htmlspecialchars($u) ?></td>
                                            <td>
                                                <form method="POST" style="display:inline" onsubmit="return confirm('Desbanir?')">
                                                    <input type="hidden" name="action" value="unban_user">
                                                    <input type="hidden" name="public_key" value="<?= htmlspecialchars($u) ?>">
                                                    <button type="submit" class="btn btn-g btn-s">Desbanir</button>
                                                </form>
                                            </td>
                                        </tr>
                                    <?php endforeach; ?>
                                </tbody>
                            </table>
                        <?php endif; ?>
                    </div>

                <?php elseif ($page === 'banned_hashes'): ?>
                    <div class="ph">
                        <h1>🔒 Hashes Banidos</h1>
                        <p>Bloqueio por SHA256</p>
                    </div>
                    <div class="card">
                        <h3>Banir Hash</h3>
                        <form method="POST">
                            <input type="hidden" name="action" value="ban_hash">
                            <div class="fg"><label>Hash SHA256</label><input type="text" name="hash" placeholder="a1b2c3..."
                                    maxlength="64" required></div>
                            <div class="fg"><label>Motivo</label><input type="text" name="reason"
                                    placeholder="Conteúdo ofensivo"></div>
                            <button type="submit" class="btn btn-d">Bloquear</button>
                        </form>
                    </div>
                    <div class="card">
                        <h3>Bloqueados (<?= count($bannedHashes) ?>)</h3>
                        <?php if (empty($bannedHashes)): ?>
                            <div class="empty">
                                <div class="ic">🔐</div>
                                <p>Nenhum</p>
                            </div>
                        <?php else: ?>
                            <table class="dt">
                                <thead>
                                    <tr>
                                        <th>Hash</th>
                                        <th>Motivo</th>
                                        <th>Ação</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <?php foreach ($bannedHashes as $h): ?>
                                        <tr>
                                            <td title="<?= htmlspecialchars($h['hash']) ?>">
                                                <?= htmlspecialchars(substr($h['hash'], 0, 16)) ?>...
                                            </td>
                                            <td style="color:var(--muted)"><?= htmlspecialchars($h['reason'] ?? '—') ?></td>
                                            <td>
                                                <form method="POST" style="display:inline" onsubmit="return confirm('Remover?')">
                                                    <input type="hidden" name="action" value="unban_hash">
                                                    <input type="hidden" name="hash" value="<?= htmlspecialchars($h['hash']) ?>">
                                                    <button type="submit" class="btn btn-g btn-s">Remover</button>
                                                </form>
                                            </td>
                                        </tr>
                                    <?php endforeach; ?>
                                </tbody>
                            </table>
                        <?php endif; ?>
                    </div>

                <?php elseif ($page === 'peers'): ?>
                    <div class="ph">
                        <h1>🌐 Peers (Gossip)</h1>
                        <p>Nós na rede</p>
                    </div>
                    <div class="card">
                        <h3>Peers (<?= count($peers) ?>)</h3>
                        <?php if (empty($peers)): ?>
                            <div class="empty">
                                <div class="ic">🌍</div>
                                <p>Nenhum peer. Configure SEED_NODES.</p>
                            </div>
                        <?php else: ?>
                            <table class="dt">
                                <thead>
                                    <tr>
                                        <th>URL</th>
                                        <th>PubKey</th>
                                        <th>Reputação</th>
                                        <th>Visto</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <?php foreach ($peers as $p): ?>
                                        <?php $rep = intval($p['reputation'] ?? 0);
                                        $cls = $rep >= 80 ? 'bg' : ($rep >= 40 ? 'bw' : 'bb'); ?>
                                        <tr>
                                            <td><a href="<?= htmlspecialchars($p['url'] ?? '') ?>" target="_blank"
                                                    style="color:var(--accent)"><?= htmlspecialchars($p['url'] ?? '—') ?></a></td>
                                            <td><?= htmlspecialchars(substr($p['public_key'] ?? '', 0, 16)) ?>...</td>
                                            <td><span class="badge <?= $cls ?>"><?= $rep ?></span></td>
                                            <td style="color:var(--muted)"><?= htmlspecialchars($p['last_seen'] ?? '—') ?></td>
                                        </tr>
                                    <?php endforeach; ?>
                                </tbody>
                            </table>
                        <?php endif; ?>
                    </div>

                <?php elseif ($page === 'delete'): ?>
                    <div class="ph">
                        <h1>🗑️ Eliminar Conteúdo</h1>
                        <p>Remover nodes por UUID</p>
                    </div>
                    <div class="card" style="border-color:var(--danger)">
                        <h3 style="color:var(--danger)">⚠️ Zona de Perigo</h3>
                        <p style="color:var(--muted);margin-bottom:15px">Ação irreversível neste nó.</p>
                        <form method="POST">
                            <input type="hidden" name="action" value="delete_node">
                            <div class="fg"><label>Node ID (UUID)</label><input type="text" name="node_id"
                                    placeholder="xxxxxxxx-xxxx-..." required></div>
                            <button type="submit" class="btn btn-d" onclick="return confirm('Irreversível. Continuar?')">🗑️
                                Eliminar</button>
                        </form>
                    </div>

                <?php elseif ($page === 'server'): ?>
                    <div class="ph">
                        <h1>⚙️ Trocar Servidor</h1>
                        <p>Conectar a outro backend</p>
                    </div>
                    <div class="card">
                        <h3>Atual: <span
                                style="color:var(--accent);font-family:'JetBrains Mono',monospace"><?= htmlspecialchars($serverUrl) ?></span>
                            <span class="dot"></span>
                        </h3>
                        <form method="POST" style="margin-top:15px">
                            <input type="hidden" name="action" value="change_server">
                            <div class="fg"><label>Novo URL</label><input type="url" name="server_url"
                                    placeholder="https://outro-no.com" required></div>
                            <button type="submit" class="btn btn-p">Trocar</button>
                        </form>
                    </div>

                <?php elseif ($page === 'identity'): ?>
                    <div class="ph">
                        <h1>🔑 Cofre .cromid</h1>
                        <p>Inspecionar identidade e usar a chave pública</p>
                    </div>

                    <?php if ($sessionPubKey): ?>
                        <div class="card" style="border-color:var(--accent)">
                            <h3>📋 Identidade da Sessão Atual</h3>
                            <div class="mg">
                                <div class="mi" style="grid-column:span 2">
                                    <div class="k">PubKey</div>
                                    <div class="vl"><?= htmlspecialchars($sessionPubKey) ?></div>
                                </div>
                                <div class="mi">
                                    <div class="k">Versão</div>
                                    <div class="vl">v<?= htmlspecialchars($_SESSION['cromid']['version'] ?? '?') ?></div>
                                </div>
                                <div class="mi">
                                    <div class="k">Trusted Source</div>
                                    <div class="vl"><?= htmlspecialchars($_SESSION['trusted_source'] ?? '—') ?></div>
                                </div>
                            </div>
                            <div style="margin-top:15px;display:flex;gap:10px;flex-wrap:wrap">
                                <button class="btn btn-p"
                                    onclick="navigator.clipboard.writeText('<?= htmlspecialchars($sessionPubKey) ?>').then(()=>alert('✅ Copiado!'))">📋
                                    Copiar PubKey</button>
                                <a href="?page=whitelist&prefill=<?= urlencode($sessionPubKey) ?>" class="btn btn-p"
                                    style="text-decoration:none">✅ Whitelist</a>
                                <a href="?page=banned_users&prefill=<?= urlencode($sessionPubKey) ?>" class="btn btn-d"
                                    style="text-decoration:none">⛔ Banir</a>
                            </div>
                        </div>
                    <?php else: ?>
                        <div class="card">
                            <div class="empty">
                                <div class="ic">🔐</div>
                                <p>Nenhum .cromid foi carregado nesta sessão.<br>Faça logout e login novamente com o ficheiro
                                    .cromid.</p>
                            </div>
                        </div>
                    <?php endif; ?>

                    <div class="card">
                        <h3>Inspecionar Outro .cromid</h3>
                        <div class="dz" id="inspect-dz" onclick="document.getElementById('inspect-file').click()">
                            <div class="ic">📂</div>
                            <p>Arraste ou clique para carregar</p>
                            <input type="file" id="inspect-file" accept=".cromid,.json" style="display:none"
                                onchange="inspectFile(this)">
                        </div>
                    </div>
                    <div id="inspect-result" style="display:none">
                        <div class="card" style="border-color:var(--warn)">
                            <h3>📋 Identidade Inspecionada</h3>
                            <div class="mg">
                                <div class="mi">
                                    <div class="k">Versão</div>
                                    <div class="vl" id="ins-ver">—</div>
                                </div>
                                <div class="mi" style="grid-column:span 2">
                                    <div class="k">PubKey</div>
                                    <div class="vl" id="ins-pk" style="font-size:.8em">—</div>
                                </div>
                                <div class="mi" style="grid-column:span 2">
                                    <div class="k">Trusted Source</div>
                                    <div class="vl" id="ins-ts">—</div>
                                </div>
                                <div class="mi">
                                    <div class="k">Vault</div>
                                    <div class="vl" id="ins-vt">—</div>
                                </div>
                            </div>
                            <div style="margin-top:15px;display:flex;gap:10px;flex-wrap:wrap">
                                <button class="btn btn-p" id="ins-copy">📋 Copiar PubKey</button>
                                <a id="ins-wl" href="#" class="btn btn-p" style="text-decoration:none">✅ Whitelist</a>
                                <a id="ins-ban" href="#" class="btn btn-d" style="text-decoration:none">⛔ Banir</a>
                            </div>
                        </div>
                        <div class="card">
                            <h3>🔍 JSON Raw</h3>
                            <pre id="ins-raw"
                                style="background:var(--input);padding:15px;border-radius:8px;font-family:'JetBrains Mono',monospace;font-size:.8em;color:var(--muted);overflow-x:auto;white-space:pre-wrap;word-break:break-all"></pre>
                        </div>
                    </div>
                    <script>
                        const idz = document.getElementById('inspect-dz');
                        idz.addEventListener('dragover', e => { e.preventDefault(); idz.classList.add('over') });
                        idz.addEventListener('dragleave', () => idz.classList.remove('over'));
                        idz.addEventListener('drop', e => { e.preventDefault(); idz.classList.remove('over'); if (e.dataTransfer.files.length) inspectParse(e.dataTransfer.files[0]) });
                        function inspectFile(i) { if (i.files.length) inspectParse(i.files[0]) }
                        function inspectParse(f) {
                            const r = new FileReader();
                            r.onload = function (e) {
                                try {
                                    const d = JSON.parse(e.target.result);
                                    document.getElementById('inspect-result').style.display = 'block';
                                    document.getElementById('ins-ver').textContent = 'v' + (d.version || '?');
                                    document.getElementById('ins-pk').textContent = d.pubKey || '—';
                                    document.getElementById('ins-ts').textContent = d.trusted_source || '(não configurado)';
                                    document.getElementById('ins-vt').textContent = d.vault ? '✅ Presente (' + d.vault.length + ' chars)' : '❌ Ausente';
                                    document.getElementById('ins-raw').textContent = JSON.stringify(d, null, 2);
                                    const pk = d.pubKey || '';
                                    document.getElementById('ins-copy').onclick = () => navigator.clipboard.writeText(pk).then(() => alert('✅ Copiado!'));
                                    document.getElementById('ins-wl').href = '?page=whitelist&prefill=' + encodeURIComponent(pk);
                                    document.getElementById('ins-ban').href = '?page=banned_users&prefill=' + encodeURIComponent(pk);
                                    document.getElementById('inspect-result').scrollIntoView({ behavior: 'smooth' });
                                } catch (err) { alert('❌ Ficheiro inválido: ' + err.message) }
                            };
                            r.readAsText(f);
                        }
                    </script>

                <?php endif; ?>

            </main>
        </div>
    <?php endif; ?>

</body>

</html>
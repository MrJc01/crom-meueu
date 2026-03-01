# CROM - MEUEU NODE v2.0

### The Sovereign Social Protocol for the Free Web

![License: AGPLv3](https://img.shields.io/badge/license-AGPLv3-blue.svg)
![Docker](https://img.shields.io/badge/docker-ready-green.svg)
![Version](https://img.shields.io/badge/version-2.0.0--sovereign-00d2ff.svg)

**Crom** é um protocolo social descentralizado onde a identidade é criptográfica (Ed25519) e o conteúdo é portável. **Meueu** é a implementação de referência de um nó da rede, com frontend multi-portal e ferramentas robustas de governança.

---

## 🏗️ Arquitetura

```
┌──────────────────────────────────────────────────────────────┐
│                    FRONTEND (HTML + JS SDK)                   │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐     │
│  │Micro │ │Insta │ │Video │ │Tube  │ │News  │ │Inbox │     │
│  │Blog  │ │Crom  │ │Feed  │ │Grid  │ │Dense │ │E2EE  │     │
│  └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘     │
│     └────────┴────────┴────┬───┴────────┴────────┘          │
│                            │                                 │
│  ┌─── JS SDK ──────────────┤                                 │
│  │ auth.js  (Vault .cromid)│                                 │
│  │ client.js (API calls)   │                                 │
│  │ checker.js (Integrity)  │                                 │
│  └─────────────────────────┘                                 │
├──────────────────────────────────────────────────────────────┤
│                     GO BACKEND (API)                          │
│  ┌────────────┐ ┌────────────┐ ┌──────────────────┐         │
│  │ /v1/publish│ │ /v1/query  │ │ /v1/sync (Gossip)│         │
│  │ (RateLimit │ │            │ │ (Peer Content)   │         │
│  │  Banlist   │ │            │ └──────────────────┘         │
│  │  Filter    │ │            │ ┌──────────────────┐         │
│  │  Whitelist)│ │            │ │ /v1/peers        │         │
│  └────────────┘ └────────────┘ └──────────────────┘         │
│  ┌────────────────────────────────────────────────┐         │
│  │ /admin/* (Token Auth: X-Admin-Token + SHA256)  │         │
│  │ stats│whitelist│ban_words│ban_users│ban_hashes │         │
│  └────────────────────────────────────────────────┘         │
├──────────────────────────────────────────────────────────────┤
│                      POSTGRESQL                              │
│  nodes │ peers │ whitelist │ banned_* │ used_nonces           │
└──────────────────────────────────────────────────────────────┘
```

---

## 🚀 Quick Start

### Requisitos
- **Go 1.23+** e **PostgreSQL** (ou Docker)
- `git`

### Dev Mode (Sem Docker)
```bash
git clone https://github.com/your-username/crom-meueu.git
cd crom-meueu

# 1. Configurar ambiente
cp .env.example .env
# Edite .env com suas credenciais do PostgreSQL

# 2. Gerar senha admin
./scripts/gen_pass.sh "minha_senha_secreta"
# Copie o hash para ADMIN_SECRET_HASH no .env

# 3. Iniciar
go run ./cmd/api
```

### Docker
```bash
docker-compose up -d
```

### Acesso
| Serviço | URL |
|---------|-----|
| **Frontend** | http://localhost:8080 |
| **Admin (JS)** | http://localhost:8080/admin.html |
| **Admin (PHP)** | http://localhost:9000 → Ver [admin/README.md](admin/README.md) |
| **API** | http://localhost:8080/v1/query |
| **Meta** | http://localhost:8080/meta |

---

## ⚙️ Configuração (.env)

```env
# Servidor
SERVER_PORT=8080
SERVER_MODE=NORMAL          # NORMAL ou WHITELIST

# Rede Descentralizada
NETWORK_ID=meueu-mainnet-v1
SERVER_URL=http://localhost:8080
SEED_NODES=                 # URLs de outros nós (vírgula)
SYNC_EXTERNAL_POSTS=true    # Aceitar posts históricos de peers

# Admin
ADMIN_SECRET_HASH=...       # SHA256 da senha admin

# Banco de Dados
DATABASE_URL=postgres://crom:crom_secret@localhost:5432/crom_db?sslmode=disable
```

---

## 🌌 Portais Frontend

Todos os portais compartilham a mesma identidade e pool de conteúdo:

| Portal | Rota | Estilo |
|--------|------|--------|
| 📢 New Post | `/publish.html` | Publicação |
| 🐦 Micro-Blog | `/twitter.html` | Timeline de texto |
| 📸 InstaCrom | `/instagram.html` | Feed visual |
| 📱 Vertical Feed | `/tiktok.html` | Scroll imersivo |
| 📺 Video Grid | `/youtube.html` | Descoberta |
| 📰 Tech News | `/tabnews.html` | Layout denso |
| 👥 Social Mix | `/facebook.html` | Layout clássico |
| 🔐 Inbox | `/inbox.html` | DMs E2E encriptadas |
| 🔍 Explorer | `/explorer.html` | Inspetor de dados |

---

## 🔑 Identidade Soberana (.cromid)

A identidade no Crom é **zero-knowledge** — a chave privada **nunca sai do dispositivo**.

### Gerar Identidade
1. Acesse qualquer portal → Clique em **"Login / Sign Up"**
2. Insira uma **passphrase** (mín. 6 chars) → O ficheiro `.cromid` é baixado
3. Guarde este ficheiro com segurança — ele **é** a sua identidade

### Estrutura do .cromid
```json
{
  "version": 1,
  "pubKey": "abc123...",
  "vault": "encrypted_aes_gcm_base64...",
  "trusted_source": "https://raw.githubusercontent.com/user/repo/main/"
}
```

- **vault**: Chave privada encriptada com AES-GCM (PBKDF2 100k iterações)
- **trusted_source** *(opcional)*: URL para verificação de integridade do frontend

### SDK JavaScript
```javascript
const auth = new CromAuth();

// Gerar nova identidade
const cromid = await auth.generateIdentity("minha_passphrase", "https://github.com/...");

// Carregar identidade existente
await auth.loadIdentity(cromidData, "minha_passphrase");

// Assinar mensagem
const signature = auth.sign(messageBytes);

// Exportar para download
auth.exportIdentity(cromidData);
```

📚 Documentação completa do SDK: [docs/frontend/GUIDE.md](docs/frontend/GUIDE.md)

---

## 👑 Administração

### Admin JS (Built-in)
Acesse `/admin.html` → Insira a senha admin (texto puro).

### Admin PHP (Avançado)
Painel visual completo com sidebar, dashboard, e gerenciamento multi-nó.

```bash
cd admin/
php -S localhost:9000
```

📚 **Guia completo**: [admin/README.md](admin/README.md)

### Funcionalidades Admin

| Recurso | Descrição | Endpoint API |
|---------|-----------|--------------|
| 📊 Dashboard | Stats do nó (nodes, users, ativos 24h) | `GET /admin/stats` |
| ✅ Whitelist | Controle de acesso por chave pública | `GET/POST/DEL /admin/whitelist` |
| ⛔ Banir User | Bloquear chave pública de publicar | `GET/POST/DEL /admin/banned_users` |
| 🚫 Palavras | Filtro de conteúdo por palavras-chave | `GET/POST/DEL /admin/banned_words` |
| 🔒 Hashes | Bloquear conteúdo por SHA256(assinatura) | `GET/POST/DEL /admin/banned_hashes` |
| 🗑️ Delete | Remover nodes específicos por UUID | `DEL /admin/node?id=` |

> Auth: Header `X-Admin-Token: <senha_em_texto>` → Backend faz SHA256 e compara.

---

## 🌐 Rede Gossip (Peer-to-Peer)

Cada nó descobre e sincroniza conteúdo com outros nós automaticamente.

### Como funciona:
1. **SEED_NODES**: Configure URLs de nós iniciais no `.env`
2. **Descoberta**: O nó faz gossip a cada 2 min, descobrindo novos peers
3. **Sync de Conteúdo**: Puxa posts recentes de peers com reputação > 50
4. **Validação**: Assinaturas Ed25519 são verificadas — posts inválidos → punição
5. **Reputação**: Peers que enviam dados inválidos perdem reputação (0 = desconectado)

### Endpoints de Rede
| Endpoint | Descrição |
|----------|-----------|
| `GET /v1/peers` | Lista de peers ativos (últimas 24h) |
| `GET /v1/sync?since=<unix>&limit=50` | Posts recentes para sincronização |
| `GET /meta` | Metadados do nó (versão, network_id, modo) |

### Controle do Operador
- `SYNC_EXTERNAL_POSTS=false` → Rejeita importações históricas (só posts frescos)
- `NETWORK_ID` → Isola redes (posts de outro network_id são ignorados)

---

## 🛡️ Segurança

### Camadas de Proteção

| Camada | Mecanismo |
|--------|-----------|
| **Anti-Replay** | Nonces únicos por autor (tabela `used_nonces`, TTL 24h) |
| **Anti-Forge** | Ed25519 + TLV canonical message (protocolo v2) |
| **Rate Limit** | 100 req/min por IP, 30 req/min por PubKey |
| **Content Filter** | Palavras banidas + Hashes banidos (SHA256) |
| **Access Control** | Whitelist + Banlist por chave pública |
| **Body Limit** | 1MB max por request |
| **Frontend Integrity** | `checker.js` verifica hashes contra `trusted_source` |

### Integridade do Frontend (checker.js)
Se o utilizador configurar `trusted_source` no `.cromid`, o `checker.js` compara os hashes SHA256 de `api.js` e `auth.js` contra um `manifest.json` remoto. Em caso de **adulteração detectada**, a chave privada é **bloqueada** e um alerta vermelho é exibido.

---

## 📂 Estrutura do Projeto

```
crom-meueu/
├── cmd/
│   ├── api/main.go          # Entry point do servidor
│   ├── crom-cli/             # CLI tool
│   └── migrate/              # Migrations runner
├── internal/
│   ├── api/
│   │   ├── handlers/         # publish, query, sync, peer, admin
│   │   └── middleware/       # banlist, whitelist, content_filter,
│   │                         # rate_limiter, auth_admin, cors
│   ├── core/
│   │   ├── domain/           # Node, Peer models
│   │   ├── security/         # Ed25519 signature verification (TLV)
│   │   └── services/         # Peer discovery + Content sync
│   └── storage/postgres/     # Repositories (node, peer, admin, nonce)
├── frontend/
│   ├── js/
│   │   ├── sdk/              # auth.js, client.js
│   │   ├── checker.js        # Frontend integrity verifier
│   │   └── api.js, ui_*.js   # UI logic
│   ├── css/                  # Stylesheets
│   └── *.html                # Portal pages (14 portais)
├── admin/                    # 📚 PHP Admin Panel
│   ├── index.php             # Painel completo
│   └── README.md             # Setup guide
├── db/migrations/            # SQL migrations (000001-000007)
├── docker-compose.yml
├── .env / .env.example
└── README.md                 # ← Você está aqui
```

---

## 📜 Licença & Transparência

Licenciado sob **AGPLv3**.

> **Cláusula de Transparência**: Não é permitido modificar secretamente a lógica do backend. Qualquer modificação no código do nó deve ser informada aos utilizadores via o endpoint `/meta`.

Consulte [LICENSE](LICENSE) para detalhes completos.

# CROM - MEUEU NODE
### The Social Protocol for the Free Web

![License: AGPLv3](https://img.shields.io/badge/license-AGPLv3-blue.svg)
![Docker](https://img.shields.io/badge/docker-ready-green.svg)
![Status](https://img.shields.io/badge/status-active-success.svg)

**Crom** is a decentralized social protocol where identity is cryptographic (Ed25519) and content is portable. **Meueu** is the reference implementation of a network node, featuring a multi-view frontend and robust governance tools.

---

## 🚀 Quick Start (Docker)
The fastest way to run your own node.

### 1. Requirements
- Docker & Docker Compose
- `git`

### 2. Run
```bash
git clone https://github.com/your-username/crom-meueu.git
cd crom-meueu

# Start the node (Background)
docker-compose up -d
```

### 3. Access
- **Frontend**: [http://localhost:8080](http://localhost:8080)
- **API**: [http://localhost:8080/v1/query](http://localhost:8080/v1/query)

---

## 🌌 User Guide
The Meueu frontend offers multiple "Portals" to view the same underlying data:

- **Micro-Blog (X-like)**: `/twitter.html` - Best for text updates.
- **Vertical Feed (Tok-like)**: `/tiktok.html` - Immersive scrolling.
- **Video Grid (Tube-like)**: `/youtube.html` - Content discovery.
- **Tech News**: `/tabnews.html` - Dense, information-rich layout.

*All portals share the same identity and content pool.*

---

## 👑 Admin Guide ("God Mode")
This node includes a powerful Governance System for server operators.

### 1. Become an Admin
To access the Admin Dashboard, you need to set a `ADMIN_SECRET_HASH` in your environment.

1. **Generate your Hash**:
   Run the included utility script:
   ```bash
   ./scripts/gen_pass.sh "my_super_secret_password"
   # Output example: 5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8
   ```

2. **Configure Node**:
   Edit `docker-compose.yml` (or `.env`):
   ```yaml
   environment:
     - ADMIN_SECRET_HASH=5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8
     - SERVER_MODE=NORMAL # or WHITELIST
   ```
   Restart: `docker-compose up -d`

3. **Login**:
   Go to [http://localhost:8080/admin.html](http://localhost:8080/admin.html) and enter your password.

### 2. Features
- **Dashboard**: View real-time node statistics.
- **Whitelist Mode**: If `SERVER_MODE=WHITELIST`, only keys added via the dashboard can post.
- **Content Filter**: Ban specific words (e.g. spam, slurs) from being posted.

---

## 🛠️ Developers

### SDK & Integration
Build your own interfaces using our JavaScript SDK.
- **Documentation**: [docs/frontend/GUIDE.md](docs/frontend/GUIDE.md)
- **Source**: `frontend/js/sdk/`

```javascript
// Example
const auth = new CromAuth();
const pubKey = await auth.login("my seed phrase");
```

### License & Transparency
This project is licensed under **AGPLv3**.
> **Transparency Clause**: You cannot secretly modify the backend logic. If you modify the node code, you must inform users via the `/meta` endpoint or connection handshake.

See [LICENSE](LICENSE) for full details.

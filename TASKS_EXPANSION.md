# TASKS_EXPANSION.md - Project Expansion Checklist

## 1. Infra (Docker & License)
- [ ] **Dockerization**
    - [ ] Create `Dockerfile` (Multistage: Go Build -> Alpine/Scratch).
    - [ ] Create `docker-compose.yml` (Postgres 15 + Go API).
    - [ ] Add `entrypoint.sh` for DB migrations if needed.
- [ ] **Legal & Governance**
    - [ ] Create `LICENSE` file (AGPLv3).
    - [ ] Add "Transparency Clause" mechanism (e.g., /status endpoint showing git hash or signed manifest).

## 2. Backend Admin (Auth & Middleware)
- [ ] **Admin Authentication**
    - [ ] Define `ADMIN_SECRET_HASH` in `.env`.
    - [ ] Middleware `RequireAdmin` specifically for `/admin/` routes.
- [ ] **Access Control (Whitelist)**
    - [ ] Add `SERVER_MODE` to config/env.
    - [ ] Middleware `CheckWhitelist`: Only allow PubKeys in `allowed_users` table to `POST`.
- [ ] **Content Moderation**
    - [ ] Create `banned_words` table or config.
    - [ ] Middleware `ContentFilter`: Scan JSON payloads for prohibited terms.
- [ ] **Admin API Endpoints**
    - [ ] `GET /admin/stats`: User counts, DB size, active nodes.
    - [ ] `POST /admin/ban`: Mark PubKey as banned.
    - [ ] `POST /admin/whitelist`: Add PubKey to whitelist.

## 3. Frontend Admin UI
- [ ] **Structure**
    - [ ] `frontend/admin.html`: Minimalist, "Hacker" aesthetic dashboard.
    - [ ] `frontend/js/admin_auth.js`: Handle Admin Token storage (SessionStorage).
- [ ] **Features**
    - [ ] **Dashboard**: Real-time stats view.
    - [ ] **User Management**: Input field to Ban/Whitelist PubKeys.
    - [ ] **Word Filter**: List editor to add/remove banned words.

## 4. SDK & Docs
- [ ] **Refactoring (The "Extraction")**
    - [ ] Isolate `IdentityManager` logic into `frontend/js/sdk/auth.js`.
    - [ ] Isolate `fetchNodes` / `publish` logic into `frontend/js/sdk/client.js`.
    - [ ] Ensure existing `ui_*.js` files import/use the new SDK structure.
- [ ] **Documentation**
    - [ ] Create `docs/frontend/GUIDE.md`:
        - [ ] "Key Generation" (Ed25519 + SHA256 Seed).
        - [ ] "Signing Protocol" (Canonical JSON format).
        - [ ] "Encryption" (X25519 + NaCl Box).
    - [ ] Update `README.md` to point to the new docs.

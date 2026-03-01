# 🛡️ Crom Admin — Painel de Controle PHP

Painel visual para gerenciar nós Crom-Meueu. Funciona como um cliente HTTP que se conecta à API do backend Go via `X-Admin-Token`.

---

## 🚀 Quick Start

### 1. Requisitos

- **PHP 8.0+** com extensões `json` e `session` (padrão na maioria das instalações)
- **Servidor backend Crom** rodando (Go API na porta 8080)

### 2. Iniciar o Painel

```bash
# Na pasta admin/
cd admin/
php -S localhost:9000
```

### 3. Acessar

Abra o navegador em: **http://localhost:9000**

### 4. Login

1. **Servidor Backend**: Insira o URL do seu nó Crom (ex: `http://localhost:8080`)
2. **Token de Admin**: Insira a **senha em texto puro** (a mesma usada no `gen_pass.sh`)
   - ⚠️ A senha é enviada via header `X-Admin-Token` para o backend
   - O backend faz `SHA256(token)` e compara com `ADMIN_SECRET_HASH`

> **Exemplo**: Se você gerou o hash com `./scripts/gen_pass.sh "123456"`, insira `123456` no campo de token.

---

## 📋 Funcionalidades

| Página | Descrição |
|--------|-----------|
| **Dashboard** | Estatísticas do nó (nodes, users, ativos 24h) + metadados (`/meta`) |
| **Peers** | Visualização dos nós da rede gossip com reputação |
| **Whitelist** | Gerenciar chaves autorizadas (modo `WHITELIST`) |
| **Utilizadores Banidos** | Banir/desbanir chaves públicas |
| **Palavras Banidas** | Filtro de conteúdo por palavras-chave |
| **Hashes Banidos** | Bloquear conteúdo por SHA256 da assinatura |
| **Eliminar Conteúdo** | Remover nodes específicos por UUID |
| **Trocar Servidor** | Conectar a outro nó backend (multi-nó) |

---

## 🔧 Multi-Node Management

O painel permite gerenciar **múltiplos nós** sem reinstalar:

1. Faça login no primeiro servidor
2. Vá em **⚙️ Trocar Servidor** na sidebar
3. Insira o URL do novo backend
4. O painel reconecta usando o mesmo token

> O token de admin pode ser diferente entre servidores. Se a conexão falhar, faça logout e login com o novo token.

---

## 🐳 Docker (Opcional)

Se quiser rodar o admin panel dentro de um container:

```dockerfile
FROM php:8.2-cli
COPY admin/ /app/admin/
WORKDIR /app/admin
EXPOSE 9000
CMD ["php", "-S", "0.0.0.0:9000"]
```

---

## 🔒 Segurança

| Ponto | Status |
|-------|--------|
| Token nunca é armazenado em disco | ✅ (sessão PHP em memória) |
| Comunicação com backend via `X-Admin-Token` | ✅ |
| Hash SHA256 no servidor | ✅ |
| HTTPS recomendado em produção | ⚠️ |

> **IMPORTANTE**: Em produção, rode este painel atrás de HTTPS (nginx/Caddy) para proteger o token em trânsito.

---

## 📂 Estrutura

```
admin/
├── index.php      # Painel completo (SPA-like com PHP routing)
└── README.md      # Este ficheiro
```

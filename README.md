# Crom Social Protocol (Draft)

**Uma infraestrutura de rede social descentralizada, agnóstica a frontend e focada na soberania dos dados.**

Este projeto ("meueu") é uma implementação de referência de um protocolo social onde o Backend serve como uma "Plataforma de Verdade" criptográfica e o Frontend é apenas uma das infinitas "Visões" possíveis sobre os dados.

## 🚀 Filosofia

*   **API First**: O produto é a API. O Frontend é apenas uma "skin".
*   **Modularidade Extrema**: O Frontend define as regras de consulta/feed. O Backend é apenas um cofre burro e seguro.
*   **Identidade Criptográfica**: Não há usuários/senhas. Tudo é assinado com chaves **Ed25519**.
*   **Agnóstico a Mídia**: Um `Node` pode ser um texto, um vídeo, uma música ou uma interação.

---

## 🛠️ Quick Start

### Pré-requisitos
*   Go 1.22+
*   Docker (para o PostgreSQL)

### 1. Iniciar o Ambiente
Utilize o script de verificação que sobe o banco, migra as tabelas e compila tudo:

```bash
./scripts/verify.sh
```

Se preferir rodar manualmente as partes:
```bash
./scripts/start_db.sh     # Sobe o Postgres no Docker
./bin/api                 # Roda a API na porta 8080
```

### 2. Popular o Banco (Seeder)
Para ver a rede social "viva", gere dados falsos:

```bash
go run scripts/seeder.go
```
*Gera ~30 posts variados (Vídeos, Textos, Artigos).*

### 3. Acessar o Portal (Frontend)
Abra o navegador em:

👉 **http://localhost:8080/**

Você verá o portal "Choose Your Reality", onde pode escolher navegar pelos dados como se estivesse no Twitter, TikTok, YouTube, etc.

---

## 📂 Estrutura do Projeto

### Backend (Go)
*   **`cmd/api`**: Entrypoint do servidor HTTP.
*   **`internal/core/domain`**: Modelos de dados (`Node`, `Filter`).
*   **`internal/core/security`**: Validação de assinaturas Ed25519.
*   **`internal/storage/postgres`**: Repositório e Query Builder dinâmico.
*   **`db/migrations`**: Scripts SQL versionados.

### Frontend (Vanilla JS)
*   **`frontend/index.html`**: Portal inicial.
*   **`frontend/*.html`**: Vistas especializadas (TikTok, Twitter, etc).
*   **`frontend/js/api.js`**: Cliente da API `POST /v1/query`.

### Ferramentas (`scripts/`)
*   **`verify.sh`**: Automação de CI/CD local.
*   **`publish_client.go`**: CLI para postar conteúdo assinado manualmente.
*   **`seeder.go`**: Gerador de massa de dados.

---

## 🔌 API Endpoints

### `POST /v1/publish`
Publica um novo conteúdo. Requer payload assinado.
*   **Body**: JSON com `author_pubkey`, `signature`, `payload`, etc.
*   **Segurança**: Verifica assinatura Ed25519 sobre a string canônica.

### `POST /v1/query`
Busca flexível estilo banco de dados.
*   **Body**:
    ```json
    {
      "filters": {
        "kinds": ["video", "text"],
        "authors": ["..."],
        "tags": ["tech"]
      },
      "limit": 20
    }
    ```

---

## 📜 Documentação Completa
Veja a pasta `docs/` para detalhes arquiteturais:
*   [01-VISION.md](docs/01-VISION.md)
*   [02-ARCHITECTURE.md](docs/02-ARCHITECTURE.md)
*   [03-PROTOCOL_API.md](docs/03-PROTOCOL_API.md)
*   [04-SECURITY_IMPORT.md](docs/04-SECURITY_IMPORT.md)
*   [05-DATABASE.md](docs/05-DATABASE.md)

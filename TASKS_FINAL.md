# Checklist Final: Consolidação & SDK Universal

Este documento define o roteiro para transformar o Crom/Meueu em um ecossistema consistente e agnóstico de plataforma.

## 1. Core Backend: Suporte a Mensageria Privada (Blind Server)
- [ ] **Auditoria de Query**: Verificar se `Save` e `Query` suportam nativamente tags JSONB para DMs.
    - [ ] Garantir indexação GIN em `tags` no Postgres.
    - [ ] Implementar filtro `@>` para queries como `{"tags": ["recipient:PUBKEY"]}`.
- [ ] **Segurança**: Confirmar que o payload é armazenado como JSONB opaco (o servidor não valida o conteúdo cifrado).

## 2. Universal Go SDK (`pkg/sdk`)
Biblioteca nativa para replicar a lógica do `frontend/js/sdk/auth.js`.
- [ ] **Auth**: `Login(seed)` -> Ed25519 Private Key + X25519 Derivation.
- [ ] **Sign**: `Sign(payload)` -> Canonical JSON String -> Ed25519 Signature.
- [ ] **Encryption**: `EncryptDM(recipient, msg)` -> Compatibilidade total com `nacl.box` (JS).
    - [ ] *Nota*: Go não tem `nacl` na stdlib, usar `golang.org/x/crypto/nacl/box` ou `secretbox`.

## 3. CLI Tool (`cmd/crom-cli`)
Ferramenta de terminal para interagir com a rede (substituindo scripts soltos).
- [ ] Implementar comandos:
    - [ ] `crom-cli login "seed phrase"` (mostra chaves).
    - [ ] `crom-cli post "Hello World"` (posta texto).
    - [ ] `crom-cli dm --to=PUBKEY "Secret Msg"` (envia DM cifrada).
    - [ ] `crom-cli inbox` (baixa e decifra DMs).

## 4. Admin & Governança (Refinamento)
- [ ] **Audit Log**:
    - [ ] Frontend: Adicionar aba "Audit Log" em `admin.html`.
    - [ ] Backend: Registrar ações de banimento/whitelist em tabela ou log estruturado.
- [ ] **Verificação**: Testar rotação de `ADMIN_SECRET_HASH`.

## 5. Documentação Final
- [ ] **Protocol Spec (`docs/SDK_SPEC.md`)**:
    - [ ] Detalhar algoritmo de derivação de chaves (Ed25519 -> X25519).
    - [ ] Especificar formato canônico para assinatura (evitar problemas de JSON key order).
    - [ ] Documentar formato do Payload de DMs (`nonce` + `ciphertext`).

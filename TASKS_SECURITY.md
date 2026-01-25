# Checklist de Segurança e Criptografia (E2EE)

## Fase 1: Criptografia Core (Frontend)
- [ ] Implementar `IdentityManager.login()` (Derivação Hash -> Seed -> KeyPair)
- [ ] Implementar `IdentityManager.getEncryptionKeys()` (Conversão Ed25519 -> X25519 via `ed2curve`)
- [ ] Implementar `encryptDirectMessage` (Box creation)
- [ ] Implementar `decryptDirectMessage` (Box opening)

## Fase 2: Interface de Usuário (Login)
- [ ] Criar `ui_login.js`: Modal que pede a "Frase Secreta".
- [ ] Integrar UI Global: Mostrar ícone/avatar com o ID curto do usuário (ex: User: aF39...) no canto da tela.
- [ ] Persistência: Manter sessão ativa na aba (SessionStorage) mas nunca salvar chave privada em LocalStorage/Cookies persistentes por segurança padrão.

## Fase 3: Mensageria (Chat)
- [ ] Criar `ui_dm.js`: Interface para listar conversas.
- [ ] Lógica de Inbox: Fazer POST `/query` com filtro `tags: ["recipient:<MEU_ID>"]`.
- [ ] Lógica de Decrypt em Massa: Ao receber a lista do servidor, tentar descriptografar cada item localmente.
- [ ] Lógica de Envio: Botão "Enviar DM" no perfil de outro usuário que cria o Node `encrypted_dm`.

## Fase 4: Backend Support
- [ ] Verificar se o Query Engine suporta filtro de tags JSONB (operador `?|` ou `@>`). Se não, ajustar `node_repository.go`.

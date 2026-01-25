# Segurança, Identidade e Migração

## Identidade (Auth)
- Não usamos "usuário/senha" tradicionais.
- A identidade é um par de chaves **Ed25519**.
- O ID do usuário (Author ID) é sua Chave Pública (ou hash dela).

## O "Paradoxo do Tempo" (Regras de Importação)
Como permitir importação de dados antigos sem permitir falsificação de histórico?

### Estrutura de Timestamp Duplo
Cada Node armazenado no banco possui dois marcos temporais:

1. **claimed_at** (Assinado pelo User): A data que o usuário diz que criou o conteúdo. Usada para ordenação visual no frontend.
2. **verified_at** (Gerado pelo Servidor): A data que o servidor viu esse dado pela primeira vez.

### Processo de Verificação
Ao importar dados de outro servidor:

1. O servidor recebe o dump de dados.
2. Valida a assinatura criptográfica de cada Node com a chave pública do autor.
3. Se a assinatura for válida, o dado é salvo com `verified_at = NOW()`.
4. O Frontend exibe um selo "Importado/Não verificado pelo servidor atual" se a diferença entre `claimed_at` e `verified_at` for grande.

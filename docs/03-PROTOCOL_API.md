# Especificação da API e Query Engine
Diferente de APIs REST tradicionais, a API do Crom Social expõe um motor de consulta configurável pelo cliente.

## Endpoint Mestre: POST /v1/query
O Frontend envia um payload definindo o que quer ver.

### Exemplo de Request (Frontend -> Backend)
```json
{
  "filters": {
    "types": ["text", "video"],
    "tags": ["dev", "physics"],
    "min_trust_score": 10,
    "date_range": { "start": "2025-01-01" }
  },
  "sort": {
    "field": "created_at",
    "order": "desc"
  },
  "limit": 50
}
```

## Endpoints de Escrita
- **POST /v1/publish**: Publica um novo Node (requer assinatura).
- **POST /v1/action**: Realiza ações (like, report) que também são Nodes de interação.

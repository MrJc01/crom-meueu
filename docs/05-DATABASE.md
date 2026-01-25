# Database Schema (PostgreSQL)

```sql
CREATE TABLE nodes (
    id UUID PRIMARY KEY,
    parent_id UUID REFERENCES nodes(id), -- Para threads/respostas
    author_pubkey VARCHAR(64) NOT NULL,
    
    -- O Conteúdo Flexível
    kind VARCHAR(20) NOT NULL, -- 'text', 'video', 'image'
    payload JSONB NOT NULL,    -- O corpo do post
    tags JSONB,                -- Indexação rápida

    -- Segurança e Tempo
    signature TEXT NOT NULL,         -- Assinatura Ed25519
    claimed_at TIMESTAMP NOT NULL,   -- Data 'dita' pelo autor
    verified_at TIMESTAMP DEFAULT NOW(), -- Data real de inserção
    origin_server VARCHAR(255),      -- De onde veio esse dado

    -- Índices para performance do Query Engine
    CONSTRAINT valid_signature CHECK (length(signature) > 0)
);

CREATE INDEX idx_nodes_payload ON nodes USING gin (payload);
CREATE INDEX idx_nodes_tags ON nodes USING gin (tags);
```

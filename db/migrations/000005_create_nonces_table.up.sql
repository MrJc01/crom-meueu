CREATE TABLE IF NOT EXISTS used_nonces (
    nonce VARCHAR(255) NOT NULL,
    author_pubkey VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    UNIQUE (nonce, author_pubkey)
);

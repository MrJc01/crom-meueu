CREATE TABLE IF NOT EXISTS banned_hashes (
    hash VARCHAR(64) PRIMARY KEY,
    reason TEXT DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

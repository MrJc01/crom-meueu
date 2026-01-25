CREATE TABLE IF NOT EXISTS banned_users (
    public_key TEXT PRIMARY KEY,
    reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_banned_users_pubkey ON banned_users(public_key);

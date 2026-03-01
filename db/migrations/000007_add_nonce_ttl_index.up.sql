-- Index for efficient TTL-based cleanup of expired nonces
CREATE INDEX IF NOT EXISTS idx_nonces_created_at ON used_nonces(created_at);

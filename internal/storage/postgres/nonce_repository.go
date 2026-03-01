package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NonceRepository struct {
	db *pgxpool.Pool
}

func NewNonceRepository(db *pgxpool.Pool) *NonceRepository {
	return &NonceRepository{db: db}
}

// MarkUsed attempts to insert the nonce. Returns true if successful, false if it already exists (replay).
func (r *NonceRepository) MarkUsed(ctx context.Context, nonce, pubKey string) bool {
	if r.db == nil { return true } // No DB = allow (fail-open in dev mode)
	query := `
		INSERT INTO used_nonces (nonce, author_pubkey, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (nonce, author_pubkey) DO NOTHING
	`
	tag, err := r.db.Exec(ctx, query, nonce, pubKey, time.Now())
	if err != nil {
		return false // Assume error means we can't accept it
	}
	return tag.RowsAffected() > 0
}

// CleanupExpired removes nonces older than the given TTL duration.
// This prevents the used_nonces table from growing infinitely.
// Recommended: call periodically (e.g., every 1 hour) with a TTL of 24-48 hours.
func (r *NonceRepository) CleanupExpired(ctx context.Context, ttl time.Duration) (int64, error) {
	if r.db == nil { return 0, nil }
	cutoff := time.Now().Add(-ttl)
	query := `DELETE FROM used_nonces WHERE created_at < $1`
	tag, err := r.db.Exec(ctx, query, cutoff)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

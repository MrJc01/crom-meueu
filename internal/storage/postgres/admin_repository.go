package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminRepository struct {
	pool *pgxpool.Pool
}

func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{pool: pool}
}

// --- Whitelist ---

func (r *AdminRepository) IsWhitelisted(ctx context.Context, pubKey string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM whitelist WHERE public_key = $1)"
	err := r.pool.QueryRow(ctx, query, pubKey).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check whitelist: %w", err)
	}
	return exists, nil
}

func (r *AdminRepository) AddToWhitelist(ctx context.Context, pubKey string) error {
	_, err := r.pool.Exec(ctx, "INSERT INTO whitelist (public_key) VALUES ($1) ON CONFLICT DO NOTHING", pubKey)
	return err
}

func (r *AdminRepository) RemoveFromWhitelist(ctx context.Context, pubKey string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM whitelist WHERE public_key = $1", pubKey)
	return err
}

func (r *AdminRepository) GetWhitelist(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, "SELECT public_key FROM whitelist ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := []string{}
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

// --- Banned Words ---

func (r *AdminRepository) GetBannedWords(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, "SELECT word FROM banned_words")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	words := []string{}
	for rows.Next() {
		var word string
		if err := rows.Scan(&word); err != nil {
			return nil, err
		}
		words = append(words, word)
	}
	return words, nil
}

// ...

func (r *AdminRepository) GetBannedUsers(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, "SELECT public_key FROM banned_users ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := []string{}
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func (r *AdminRepository) AddBannedWord(ctx context.Context, word string) error {
	_, err := r.pool.Exec(ctx, "INSERT INTO banned_words (word) VALUES ($1) ON CONFLICT DO NOTHING", word)
	return err
}

func (r *AdminRepository) RemoveBannedWord(ctx context.Context, word string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM banned_words WHERE word = $1", word)
	return err
}

// --- Stats ---
type SystemStats struct {
	TotalNodes     int64 `json:"total_nodes"`
	TotalUsers     int64 `json:"total_users"` // approx (unique authors)
	ActiveNodes24H int64 `json:"active_nodes_24h"`
}

func (r *AdminRepository) GetStats(ctx context.Context) (*SystemStats, error) {
	stats := &SystemStats{}

	// Total Nodes
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM nodes").Scan(&stats.TotalNodes); err != nil {
		return nil, err
	}

	// Unique Authors (Total Users)
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(DISTINCT author_pubkey) FROM nodes").Scan(&stats.TotalUsers); err != nil {
		return nil, err
	}

	// Active Nodes (last 24h)
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM nodes WHERE claimed_at > NOW() - INTERVAL '24 hours'").Scan(&stats.ActiveNodes24H); err != nil {
		return nil, err
	}

	return stats, nil
}

// --- Banned Users ---

func (r *AdminRepository) IsBannedUser(ctx context.Context, pubKey string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM banned_users WHERE public_key = $1)"
	err := r.pool.QueryRow(ctx, query, pubKey).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check ban: %w", err)
	}
	return exists, nil
}

func (r *AdminRepository) BanUser(ctx context.Context, pubKey, reason string) error {
	_, err := r.pool.Exec(ctx, "INSERT INTO banned_users (public_key, reason) VALUES ($1, $2) ON CONFLICT DO NOTHING", pubKey, reason)
	return err
}

func (r *AdminRepository) UnbanUser(ctx context.Context, pubKey string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM banned_users WHERE public_key = $1", pubKey)
	return err
}

// --- Content Moderation ---

func (r *AdminRepository) DeleteNode(ctx context.Context, nodeID string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM nodes WHERE id = $1", nodeID)
	return err
}

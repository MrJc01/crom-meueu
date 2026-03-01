package postgres

import (
	"context"
	"errors"
	"time"

	"meueu/internal/core/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PeerRepository struct {
	db *pgxpool.Pool
}

func NewPeerRepository(db *pgxpool.Pool) *PeerRepository {
	return &PeerRepository{db: db}
}

func (r *PeerRepository) IsAvailable() bool {
	return r.db != nil
}

func (r *PeerRepository) Upsert(ctx context.Context, p *domain.Peer) error {
	query := `
		INSERT INTO peers (id, url, public_key, last_seen, reputation)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (public_key) DO UPDATE 
		SET url = EXCLUDED.url, last_seen = EXCLUDED.last_seen
	`
	_, err := r.db.Exec(ctx, query, p.ID, p.URL, p.PublicKey, p.LastSeen, p.Reputation)
	return err
}

func (r *PeerRepository) ListActive(ctx context.Context, cutoff time.Time, limit int) ([]*domain.Peer, error) {
	query := `
		SELECT id, url, public_key, last_seen, reputation 
		FROM peers 
		WHERE last_seen >= $1 AND reputation > 0
		ORDER BY last_seen DESC 
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, query, cutoff, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peers []*domain.Peer
	for rows.Next() {
		var p domain.Peer
		if err := rows.Scan(&p.ID, &p.URL, &p.PublicKey, &p.LastSeen, &p.Reputation); err != nil {
			return nil, err
		}
		peers = append(peers, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return peers, nil
}

func (r *PeerRepository) Punish(ctx context.Context, pubKey string, penalty int) error {
	query := `UPDATE peers SET reputation = reputation - $1 WHERE public_key = $2`
	_, err := r.db.Exec(ctx, query, penalty, pubKey)
	return err
}

func (r *PeerRepository) GetByURL(ctx context.Context, url string) (*domain.Peer, error) {
	query := `SELECT id, url, public_key, last_seen, reputation FROM peers WHERE url = $1 LIMIT 1`
	var p domain.Peer
	err := r.db.QueryRow(ctx, query, url).Scan(&p.ID, &p.URL, &p.PublicKey, &p.LastSeen, &p.Reputation)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &p, nil
}

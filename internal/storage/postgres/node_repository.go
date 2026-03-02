package postgres

import (
	"context"
	"fmt"
	"time"

	"meueu/internal/core/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NodeRepository struct {
	pool *pgxpool.Pool
}

func NewNodeRepository(pool *pgxpool.Pool) *NodeRepository {
	return &NodeRepository{pool: pool}
}

// IsAvailable returns true if the database connection pool is initialized.
func (r *NodeRepository) IsAvailable() bool {
	return r.pool != nil
}

func (r *NodeRepository) Create(ctx context.Context, node *domain.Node) error {
	if r.pool == nil {
		return fmt.Errorf("database not available")
	}
	query := `
		INSERT INTO nodes (
			id, parent_id, author_pubkey, kind, payload, tags, 
			signature, claimed_at, verified_at, origin_server
		) VALUES (
			$1, $2, $3, $4, $5, $6, 
			$7, $8, NOW(), $9
		)
	`

	_, err := r.pool.Exec(ctx, query,
		node.ID,
		node.ParentID,
		node.AuthorPubkey,
		node.Kind,
		node.Payload,
		node.Tags,
		node.Signature,
		node.ClaimedAt,
		node.OriginServer,
	)

	if err != nil {
		return fmt.Errorf("failed to insert node: %w", err)
	}

	return nil
}

// CreateBatch inserts multiple nodes inside a single transaction.
// Fails and rolls back entirely if any node is malformed.
func (r *NodeRepository) CreateBatch(ctx context.Context, nodes []*domain.Node) error {
	if r.pool == nil {
		return fmt.Errorf("database not available")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO nodes (
			id, parent_id, author_pubkey, kind, payload, tags, 
			signature, claimed_at, verified_at, origin_server
		) VALUES (
			$1, $2, $3, $4, $5, $6, 
			$7, $8, NOW(), $9
		) ON CONFLICT (id) DO NOTHING
	`

	for _, node := range nodes {
		_, err := tx.Exec(ctx, query,
			node.ID,
			node.ParentID,
			node.AuthorPubkey,
			node.Kind,
			node.Payload,
			node.Tags,
			node.Signature,
			node.ClaimedAt,
			node.OriginServer,
		)
		if err != nil {
			return fmt.Errorf("batch insert failed at node %s: %w", node.ID, err)
		}
	}

	return tx.Commit(ctx)
}

func (r *NodeRepository) Query(ctx context.Context, filter domain.NodeFilter, limit, offset int) ([]*domain.Node, error) {
	if r.pool == nil {
		return []*domain.Node{}, nil // Dev mode: return empty
	}
	// Base Query
	query := `SELECT 
		id, parent_id, author_pubkey, kind, payload, tags, 
		signature, claimed_at, verified_at, origin_server
	FROM nodes WHERE 1=1`

	args := []interface{}{}
	argCounter := 1

	// Dynamic Filters
	if len(filter.IDs) > 0 {
		query += fmt.Sprintf(" AND id = ANY($%d)", argCounter)
		args = append(args, filter.IDs)
		argCounter++
	}

	if len(filter.Kinds) > 0 {
		query += fmt.Sprintf(" AND kind = ANY($%d)", argCounter)
		args = append(args, filter.Kinds)
		argCounter++
	}

	if len(filter.Authors) > 0 {
		query += fmt.Sprintf(" AND author_pubkey = ANY($%d)", argCounter)
		args = append(args, filter.Authors)
		argCounter++
	}

	if len(filter.Tags) > 0 {
		// Uses Postgres JSONB operator ?| (exists any)
		query += fmt.Sprintf(" AND tags ?| $%d", argCounter)
		args = append(args, filter.Tags)
		argCounter++
	}

	// Ordering and Pagination
	query += " ORDER BY claimed_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCounter)
		args = append(args, limit)
		argCounter++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCounter)
		args = append(args, offset)
		argCounter++
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var nodes []*domain.Node
	for rows.Next() {
		node := &domain.Node{}
		err := rows.Scan(
			&node.ID,
			&node.ParentID,
			&node.AuthorPubkey,
			&node.Kind,
			&node.Payload,
			&node.Tags,
			&node.Signature,
			&node.ClaimedAt,
			&node.VerifiedAt,
			&node.OriginServer,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (r *NodeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Node, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("database not available")
	}
	query := `SELECT 
		id, parent_id, author_pubkey, kind, payload, tags, 
		signature, claimed_at, verified_at, origin_server
	FROM nodes WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, id)

	node := &domain.Node{}
	err := row.Scan(
		&node.ID,
		&node.ParentID,
		&node.AuthorPubkey,
		&node.Kind,
		&node.Payload,
		&node.Tags,
		&node.Signature,
		&node.ClaimedAt,
		&node.VerifiedAt,
		&node.OriginServer,
	)
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	return node, nil
}

// ListRecent returns nodes with claimed_at > since, ordered by claimed_at ASC.
// Used by the /v1/sync endpoint for peer content synchronization.
func (r *NodeRepository) ListRecent(ctx context.Context, since time.Time, limit int) ([]*domain.Node, error) {
	if r.pool == nil {
		return []*domain.Node{}, nil // Dev mode: return empty
	}
	query := `SELECT 
		id, parent_id, author_pubkey, kind, payload, tags, 
		signature, claimed_at, verified_at, origin_server
	FROM nodes 
	WHERE claimed_at > $1 
	ORDER BY claimed_at ASC 
	LIMIT $2`

	rows, err := r.pool.Query(ctx, query, since, limit)
	if err != nil {
		return nil, fmt.Errorf("list recent failed: %w", err)
	}
	defer rows.Close()

	var nodes []*domain.Node
	for rows.Next() {
		node := &domain.Node{}
		err := rows.Scan(
			&node.ID,
			&node.ParentID,
			&node.AuthorPubkey,
			&node.Kind,
			&node.Payload,
			&node.Tags,
			&node.Signature,
			&node.ClaimedAt,
			&node.VerifiedAt,
			&node.OriginServer,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

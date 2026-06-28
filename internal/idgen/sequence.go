package idgen

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SequenceGenerator uses a PostgreSQL sequence for unique, concurrent-safe IDs.
type SequenceGenerator struct {
	pool *pgxpool.Pool
}

func NewSequenceGenerator(pool *pgxpool.Pool) *SequenceGenerator {
	return &SequenceGenerator{pool: pool}
}

func (g *SequenceGenerator) Next(ctx context.Context) (uint64, error) {
	var id uint64
	err := g.pool.QueryRow(ctx, "SELECT nextval('url_mapping_id_seq')").Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("next sequence id: %w", err)
	}
	return id, nil
}

package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hiiamhip/nanoflux/internal/domain"
)

type clickRepo struct {
	db *pgxpool.Pool
}

func NewClickRepo(pool *pgxpool.Pool) *clickRepo {
	return &clickRepo{db: pool}
}

func (r *clickRepo) SaveBatch(ctx context.Context, clicks []*domain.Click) error {
	if len(clicks) == 0 {
		return nil
	}

	rows := make([][]any, len(clicks))
	for i, c := range clicks {
		rows[i] = []any{c.ID, c.ShortCode, c.IPAddress, c.UserAgent, c.ClickedAt}
	}
	_, err := r.db.CopyFrom(
		ctx,
		pgx.Identifier{"clicks"},
		[]string{"id", "short_code", "ip_address", "user_agent", "clicked_at"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("copy from clicks: %w", err)
	}
	return nil
}

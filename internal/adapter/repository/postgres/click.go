package postgres

import (
	"context"

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
	return nil
}

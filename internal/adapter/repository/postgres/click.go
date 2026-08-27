package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hiiamhip/nanoflux/internal/domain"
)

type ClickRepo struct {
	db *pgxpool.Pool
}

func NewClickRepo(pool *pgxpool.Pool) *ClickRepo {
	return &ClickRepo{db: pool}
}

func (r *ClickRepo) SaveBatch(ctx context.Context, clicks []*domain.Click) error {
	return nil
}

package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hiiamhip/nanoflux/internal/domain"
)

type postgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRepo(pool *pgxpool.Pool) *postgresRepo {
	return &postgresRepo{db: pool}
}

func (r *postgresRepo) Create(ctx context.Context, url *domain.URL) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO urls (id, original, short_code, created_at, updated_at) VALUES ($1,$2,$3,$4,$5)`,
		url.ID, url.Original, url.ShortCode, url.CreatedAt, url.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("exec insert: %w", err)
	}
	return nil
}

func (*postgresRepo) FindByShortCode(ctx context.Context, code string) (*domain.URL, error) {
	panic("not implemented")
}

func (*postgresRepo) FindByID(ctx context.Context, id string) (*domain.URL, error) {
	panic("not implemented")
}

func (*postgresRepo)Update(ctx context.Context, url *domain.URL) error {
	panic("not implemented")
}

func (*postgresRepo)Delete(ctx context.Context, id string) error {
	panic("not implemented")
}

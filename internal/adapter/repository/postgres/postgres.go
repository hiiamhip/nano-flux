package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hiiamhip/nanoflux/internal/domain"
	"github.com/hiiamhip/nanoflux/internal/usecase"
)

type postgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRepo(connStr string) (usecase.URLRepository, error) {
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("pgxpool new: %w", err)
	}
	return &postgresRepo{db: pool}, nil
}

func (*postgresRepo) Create(ctx context.Context, url *domain.URL) error {
	panic("not implemented")
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

package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
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

func (r *postgresRepo) FindByShortCode(ctx context.Context, code string) (*domain.URL, error) {
	var url domain.URL
	err := r.db.QueryRow(ctx,
		`SELECT id, original, short_code, created_at, updated_at FROM urls WHERE short_code = $1`,
		code,
	).Scan(&url.ID, &url.Original, &url.ShortCode, &url.CreatedAt, &url.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query find by short code: %w", err)
	}
	return &url, nil
}

func (r *postgresRepo) FindByID(ctx context.Context, id string) (*domain.URL, error) {
	var url domain.URL
	err := r.db.QueryRow(ctx,
		`SELECT id, original, short_code, created_at, updated_at FROM urls WHERE id = $1`,
		id,
	).Scan(&url.ID, &url.Original, &url.ShortCode, &url.CreatedAt, &url.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query find by id: %w", err)
	}
	return &url, nil
}

func (r *postgresRepo) Update(ctx context.Context, url *domain.URL) error {
	_, err := r.db.Exec(ctx,
		`UPDATE urls SET original = $1, updated_at = $2 WHERE id = $3`,
		url.Original, url.UpdatedAt, url.ID,
	)
	if err != nil {
		return fmt.Errorf("exec update: %w", err)
	}
	return nil
}

func (r *postgresRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM urls WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("exec delete: %w", err)
	}
	return nil
}

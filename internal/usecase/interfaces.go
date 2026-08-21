package usecase

import (
	"context"

	"github.com/hiiamhip/nanoflux/internal/domain"
)


type URLRepository interface {
    Create(ctx context.Context, url *domain.URL) error

    FindByShortCode(ctx context.Context, code string) (*domain.URL, error)

    FindByID(ctx context.Context, id string) (*domain.URL, error)

    Update(ctx context.Context, url *domain.URL) error

    Delete(ctx context.Context, id string) error
}

type Shortener interface {
    Generate(ctx context.Context, originalURL string) (string, error)
}

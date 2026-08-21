package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hiiamhip/nanoflux/internal/domain"
)

type URLService struct {
    repo      URLRepository
    shortener Shortener
}

// NewURLService constructs a new service with its dependencies.
func NewURLService(repo URLRepository, shortener Shortener) *URLService {
    return &URLService{repo: repo, shortener: shortener}
}

// CreateShortURL validates the original URL, generates a short code, and persists the mapping.
func (s *URLService) CreateShortURL(ctx context.Context, original string) (string, error) {
    if original == "" {
        return "", domain.ErrInvalidURL
    }
    code, err := s.shortener.Generate(ctx, original)
    if err != nil {
        return "", fmt.Errorf("shortener generate: %w", err)
    }
    url := &domain.URL{
        ID:        uuid.NewString(),
        Original:  original,
        ShortCode: code,
        CreatedAt: time.Now().Unix(),
        UpdatedAt: time.Now().Unix(),
    }
    if err := s.repo.Create(ctx, url); err != nil {
        return "", fmt.Errorf("repo create: %w", err)
    }
    return code, nil
}

// ResolveURL fetches the original URL for a given short code.
func (s *URLService) ResolveURL(ctx context.Context, code string) (string, error) {
    url, err := s.repo.FindByShortCode(ctx, code)
    if err != nil {
        return "", fmt.Errorf("repo find: %w", err)
    }
    if url == nil {
        return "", domain.ErrNotFound
    }
    return url.Original, nil
}

// UpdateURL allows updating the original URL for an existing short code.
func (s *URLService) UpdateURL(ctx context.Context, id string, newOriginal string) error {
    if newOriginal == "" {
        return domain.ErrInvalidURL
    }
    url, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return fmt.Errorf("repo find by id: %w", err)
    }
    if url == nil {
        return domain.ErrNotFound
    }
    url.Original = newOriginal
    url.UpdatedAt = time.Now().Unix()
    if err := s.repo.Update(ctx, url); err != nil {
        return fmt.Errorf("repo update: %w", err)
    }
    return nil
}

// DeleteURL removes a URL mapping by its ID.
func (s *URLService) DeleteURL(ctx context.Context, id string) error {
    if err := s.repo.Delete(ctx, id); err != nil {
        return fmt.Errorf("repo delete: %w", err)
    }
    return nil
}

package usecase

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/hiiamhip/nanoflux/internal/domain"
)

type URLService struct {
    repo      URLRepository
    shortener Shortener
    tracker   ClickTracker
}

func NewURLService(repo URLRepository, shortener Shortener, tracker ClickTracker) *URLService {
    return &URLService{
        repo:      repo,
        shortener: shortener,
        tracker:   tracker,
    }
}

func (s *URLService) CreateShortURL(ctx context.Context, original string) (string, error) {
	parsed, err := url.Parse(original)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") {
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
    return s.repo.Create(ctx, url)
}

func (s *URLService) ResolveURL(ctx context.Context, code string, ip string, userAgent string) (string, error) {
	url, err := s.repo.FindByShortCode(ctx, code)
	if err != nil {
		return "", fmt.Errorf("repo find: %w", err)
	}

    click := &domain.Click{
        ID:        uuid.NewString(),
        ShortCode: code,
        IPAddress: ip,
        UserAgent: userAgent,
        ClickedAt: time.Now().Unix(),
    }
    _ = s.tracker.Track(ctx, click)

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

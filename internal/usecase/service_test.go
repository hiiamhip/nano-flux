package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/hiiamhip/nanoflux/internal/domain"
	"github.com/google/uuid"
)

type mockRepo struct {
	urls        map[string]*domain.URL
	createFunc  func(ctx context.Context, url *domain.URL) (string, error)
}

func newMockRepo() *mockRepo {
	m := &mockRepo{urls: make(map[string]*domain.URL)}
	m.createFunc = m.defaultCreate
	return m
}

func (m *mockRepo) defaultCreate(ctx context.Context, url *domain.URL) (string, error) {
	if existing, ok := m.urls[url.Original]; ok {
		return existing.ShortCode, nil
	}
	m.urls[url.ShortCode] = url
	return url.ShortCode, nil
}

func (m *mockRepo) Create(ctx context.Context, url *domain.URL) (string, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, url)
	}
	return m.defaultCreate(ctx, url)
}

func (m *mockRepo) FindByShortCode(ctx context.Context, code string) (*domain.URL, error) {
	url, ok := m.urls[code]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return url, nil
}

func (m *mockRepo) FindByID(ctx context.Context, id string) (*domain.URL, error) {
	for _, u := range m.urls {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockRepo) Update(ctx context.Context, url *domain.URL) error {
	if _, ok := m.urls[url.ShortCode]; !ok {
		return domain.ErrNotFound
	}
	m.urls[url.ShortCode] = url
	return nil
}

func (m *mockRepo) Delete(ctx context.Context, id string) error {
	for code, u := range m.urls {
		if u.ID == id {
			delete(m.urls, code)
			return nil
		}
	}
	return domain.ErrNotFound
}

type mockShortener struct {
	genFunc func(ctx context.Context, originalURL string) (string, error)
}

func (m *mockShortener) Generate(ctx context.Context, originalURL string) (string, error) {
	if m.genFunc != nil {
		return m.genFunc(ctx, originalURL)
	}
	return "", nil
}

func TestURLService_CreateShortURL_Success(t *testing.T) {
	repo := newMockRepo()
	shortener := &mockShortener{genFunc: func(ctx context.Context, originalURL string) (string, error) {
		return "abc123", nil
	}}
	svc := NewURLService(repo, shortener)

	code, err := svc.CreateShortURL(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != "abc123" {
		t.Fatalf("expected code abc123, got %s", code)
	}

	url, _ := repo.FindByShortCode(context.Background(), "abc123")
	if url.Original != "https://example.com" {
		t.Fatalf("original URL mismatch")
	}
}

func TestURLService_CreateShortURL_InvalidURL(t *testing.T) {
	repo := newMockRepo()
	shortener := &mockShortener{genFunc: func(ctx context.Context, originalURL string) (string, error) {
		return "abc123", nil
	}}
	svc := NewURLService(repo, shortener)

	_, err := svc.CreateShortURL(context.Background(), "not-a-url")
	if !errors.Is(err, domain.ErrInvalidURL) {
		t.Fatalf("expected ErrInvalidURL, got %v", err)
	}

	_, err = svc.CreateShortURL(context.Background(), "ftp://example.com")
	if !errors.Is(err, domain.ErrInvalidURL) {
		t.Fatalf("expected ErrInvalidURL for ftp, got %v", err)
	}
}

func TestURLService_CreateShortURL_ShortenerError(t *testing.T) {
	repo := newMockRepo()
	shortener := &mockShortener{genFunc: func(ctx context.Context, originalURL string) (string, error) {
		return "", errors.New("shortener failed")
	}}
	svc := NewURLService(repo, shortener)

	_, err := svc.CreateShortURL(context.Background(), "https://example.com")
	if err == nil {
		t.Fatal("expected error from shortener")
	}
}

func TestURLService_CreateShortURL_DuplicateOriginalReturnsExistingCode(t *testing.T) {
	repo := newMockRepo()
	shortener := &mockShortener{genFunc: func(ctx context.Context, originalURL string) (string, error) {
		return "abc123", nil
	}}
	svc := NewURLService(repo, shortener)

	// First call
	code1, err := svc.CreateShortURL(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}

	// Second call with same URL - should return same code
	code2, err := svc.CreateShortURL(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}

	if code1 != code2 {
		t.Fatalf("expected same code for duplicate URL: got %s and %s", code1, code2)
	}
}

func TestURLService_ResolveURL_Success(t *testing.T) {
	repo := newMockRepo()
	url := &domain.URL{
		ID:        uuid.NewString(),
		Original:  "https://example.com",
		ShortCode: "abc123",
	}
	repo.urls["abc123"] = url

	shortener := &mockShortener{}
	svc := NewURLService(repo, shortener)

	original, err := svc.ResolveURL(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if original != "https://example.com" {
		t.Fatalf("expected original URL, got %s", original)
	}
}

func TestURLService_ResolveURL_NotFound(t *testing.T) {
	repo := newMockRepo()
	shortener := &mockShortener{}
	svc := NewURLService(repo, shortener)

	_, err := svc.ResolveURL(context.Background(), "nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestURLService_UpdateURL_Success(t *testing.T) {
	repo := newMockRepo()
	url := &domain.URL{
		ID:        uuid.NewString(),
		Original:  "https://old.com",
		ShortCode: "abc123",
	}
	repo.urls["abc123"] = url

	shortener := &mockShortener{}
	svc := NewURLService(repo, shortener)

	err := svc.UpdateURL(context.Background(), url.ID, "https://new.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := repo.FindByShortCode(context.Background(), "abc123")
	if updated.Original != "https://new.com" {
		t.Fatalf("URL not updated")
	}
}

func TestURLService_UpdateURL_NotFound(t *testing.T) {
	repo := newMockRepo()
	shortener := &mockShortener{}
	svc := NewURLService(repo, shortener)

	err := svc.UpdateURL(context.Background(), "nonexistent", "https://new.com")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestURLService_UpdateURL_InvalidURL(t *testing.T) {
	repo := newMockRepo()
	shortener := &mockShortener{}
	svc := NewURLService(repo, shortener)

	err := svc.UpdateURL(context.Background(), "some-id", "")
	if !errors.Is(err, domain.ErrInvalidURL) {
		t.Fatalf("expected ErrInvalidURL, got %v", err)
	}
}

func TestURLService_DeleteURL_Success(t *testing.T) {
	repo := newMockRepo()
	url := &domain.URL{
		ID:        uuid.NewString(),
		Original:  "https://example.com",
		ShortCode: "abc123",
	}
	repo.urls["abc123"] = url

	shortener := &mockShortener{}
	svc := NewURLService(repo, shortener)

	err := svc.DeleteURL(context.Background(), url.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := repo.FindByShortCode(context.Background(), "abc123"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("URL should be deleted")
	}
}

func TestURLService_DeleteURL_NotFound(t *testing.T) {
	repo := newMockRepo()
	shortener := &mockShortener{}
	svc := NewURLService(repo, shortener)

	err := svc.DeleteURL(context.Background(), "nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
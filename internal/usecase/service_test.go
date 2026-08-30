package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/hiiamhip/nanoflux/internal/domain"
)

type mockRepo struct {
	urls       map[string]*domain.URL
	createFunc func(ctx context.Context, url *domain.URL) (string, error)
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

type mockTracker struct {
	trackFunc func(ctx context.Context, click *domain.Click) error
}

func (m *mockTracker) Track(ctx context.Context, click *domain.Click) error {
	if m.trackFunc != nil {
		return m.trackFunc(ctx, click)
	}
	return nil
}

func TestURLService_CreateShortURL_Success(t *testing.T) {
	repo := newMockRepo()
	shortener := &mockShortener{genFunc: func(ctx context.Context, originalURL string) (string, error) {
		return "abc123", nil
	}}
	tracker := &mockTracker{}
	svc := NewURLService(repo, shortener, tracker)

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
	tracker := &mockTracker{}
	svc := NewURLService(repo, shortener, tracker)

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
	tracker := &mockTracker{}
	svc := NewURLService(repo, shortener, tracker)

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
	tracker := &mockTracker{}
	svc := NewURLService(repo, shortener, tracker)

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

	var trackedClick *domain.Click
	tracker := &mockTracker{
		trackFunc: func(ctx context.Context, click *domain.Click) error {
			trackedClick = click
			return nil
		},
	}
	shortener := &mockShortener{}
	svc := NewURLService(repo, shortener, tracker)

	original, err := svc.ResolveURL(context.Background(), "abc123", "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if original != "https://example.com" {
		t.Fatalf("expected original URL, got %s", original)
	}

	if trackedClick == nil {
		t.Fatal("expected click to be tracked")
	}
	if trackedClick.ShortCode != "abc123" || trackedClick.IPAddress != "192.168.1.1" || trackedClick.UserAgent != "Mozilla/5.0" {
		t.Fatalf("tracked click mismatch: %+v", trackedClick)
	}
}

func TestURLService_ResolveURL_NotFound(t *testing.T) {
	repo := newMockRepo()
	shortener := &mockShortener{}
	tracker := &mockTracker{}
	svc := NewURLService(repo, shortener, tracker)

	_, err := svc.ResolveURL(context.Background(), "nonexistent", "127.0.0.1", "curl/7.68.0")
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
	tracker := &mockTracker{}
	svc := NewURLService(repo, shortener, tracker)

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
	tracker := &mockTracker{}
	svc := NewURLService(repo, shortener, tracker)

	err := svc.UpdateURL(context.Background(), "nonexistent", "https://new.com")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestURLService_UpdateURL_InvalidURL(t *testing.T) {
	repo := newMockRepo()
	shortener := &mockShortener{}
	tracker := &mockTracker{}
	svc := NewURLService(repo, shortener, tracker)

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
	tracker := &mockTracker{}
	svc := NewURLService(repo, shortener, tracker)

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
	tracker := &mockTracker{}
	svc := NewURLService(repo, shortener, tracker)

	err := svc.DeleteURL(context.Background(), "nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
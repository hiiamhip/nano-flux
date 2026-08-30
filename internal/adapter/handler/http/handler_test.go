package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/hiiamhip/nanoflux/internal/domain"
)

type mockService struct {
	createFunc  func(ctx context.Context, original string) (string, error)
	resolveFunc func(ctx context.Context, code, ip, userAgent string) (string, error)
	updateFunc  func(ctx context.Context, id, newOriginal string) error
	deleteFunc  func(ctx context.Context, id string) error
}

func (m *mockService) CreateShortURL(ctx context.Context, original string) (string, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, original)
	}
	return "", nil
}

func (m *mockService) ResolveURL(ctx context.Context, code, ip, userAgent string) (string, error) {
	if m.resolveFunc != nil {
		return m.resolveFunc(ctx, code, ip, userAgent)
	}
	return "", nil
}

func (m *mockService) UpdateURL(ctx context.Context, id, newOriginal string) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, newOriginal)
	}
	return nil
}

func (m *mockService) DeleteURL(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func TestHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(*mockService)
		wantStatus int
		wantCode   string
		wantError  string
	}{
		{
			name: "success",
			body: `{"url": "https://example.com"}`,
			setup: func(m *mockService) {
				m.createFunc = func(ctx context.Context, original string) (string, error) {
					return "abc123", nil
				}
			},
			wantStatus: http.StatusCreated,
			wantCode:   "abc123",
		},
		{
			name:       "invalid json",
			body:       `invalid`,
			setup:      func(m *mockService) {},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid json body",
		},
		{
			name: "missing url field",
			body: `{}`,
			setup: func(m *mockService) {
				m.createFunc = func(ctx context.Context, original string) (string, error) {
					if original == "" {
						return "", domain.ErrInvalidURL
					}
					return "abc123", nil
				}
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid url",
		},
		{
			name: "invalid url",
			body: `{"url": "not-a-url"}`,
			setup: func(m *mockService) {
				m.createFunc = func(ctx context.Context, original string) (string, error) {
					return "", domain.ErrInvalidURL
				}
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid url",
		},
		{
			name: "internal error",
			body: `{"url": "https://example.com"}`,
			setup: func(m *mockService) {
				m.createFunc = func(ctx context.Context, original string) (string, error) {
					return "", errors.New("db error")
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantError:  "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockService{}
			tt.setup(mock)
			h := NewHandler(mock)
			r := NewRouter(h)

			req := httptest.NewRequest(http.MethodPost, "/api/urls", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var resp createResponse
				json.NewDecoder(w.Body).Decode(&resp)
				if resp.Code != tt.wantCode {
					t.Errorf("code = %s, want %s", resp.Code, tt.wantCode)
				}
			}

			if tt.wantError != "" {
				var resp map[string]string
				json.NewDecoder(w.Body).Decode(&resp)
				if resp["error"] != tt.wantError {
					t.Errorf("error = %s, want %s", resp["error"], tt.wantError)
				}
			}
		})
	}
}

func TestHandler_Resolve(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		setup      func(*mockService)
		wantStatus int
		wantLoc    string
	}{
		{
			name: "success",
			code: "abc123",
			setup: func(m *mockService) {
				m.resolveFunc = func(ctx context.Context, code, ip, userAgent string) (string, error) {
					return "https://example.com", nil
				}
			},
			wantStatus: http.StatusFound,
			wantLoc:    "https://example.com",
		},
		{
			name: "not found",
			code: "nonexistent",
			setup: func(m *mockService) {
				m.resolveFunc = func(ctx context.Context, code, ip, userAgent string) (string, error) {
					return "", domain.ErrNotFound
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "internal error",
			code: "abc123",
			setup: func(m *mockService) {
				m.resolveFunc = func(ctx context.Context, code, ip, userAgent string) (string, error) {
					return "", errors.New("db error")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockService{}
			tt.setup(mock)
			h := NewHandler(mock)
			r := NewRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/"+tt.code, nil)
			ctx := chi.NewRouteContext()
			ctx.URLParams.Add("code", tt.code)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantLoc != "" {
				loc := w.Header().Get("Location")
				if loc != tt.wantLoc {
					t.Errorf("Location = %s, want %s", loc, tt.wantLoc)
				}
			}
		})
	}
}

func TestHandler_Update(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		body       string
		setup      func(*mockService)
		wantStatus int
	}{
		{
			name: "success",
			id:   "url-id-1",
			body: `{"url": "https://new.com"}`,
			setup: func(m *mockService) {
				m.updateFunc = func(ctx context.Context, id, newOriginal string) error {
					return nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid json",
			id:         "url-id-1",
			body:       `invalid`,
			setup:      func(m *mockService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty url",
			id:   "url-id-1",
			body: `{"url": ""}`,
			setup: func(m *mockService) {
				m.updateFunc = func(ctx context.Context, id, newOriginal string) error {
					return domain.ErrInvalidURL
				}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "not found",
			id:   "nonexistent",
			body: `{"url": "https://new.com"}`,
			setup: func(m *mockService) {
				m.updateFunc = func(ctx context.Context, id, newOriginal string) error {
					return domain.ErrNotFound
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "internal error",
			id:   "url-id-1",
			body: `{"url": "https://new.com"}`,
			setup: func(m *mockService) {
				m.updateFunc = func(ctx context.Context, id, newOriginal string) error {
					return errors.New("db error")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockService{}
			tt.setup(mock)
			h := NewHandler(mock)
			r := NewRouter(h)

			req := httptest.NewRequest(http.MethodPut, "/api/urls/"+tt.id, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			ctx := chi.NewRouteContext()
			ctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		setup      func(*mockService)
		wantStatus int
	}{
		{
			name: "success",
			id:   "url-id-1",
			setup: func(m *mockService) {
				m.deleteFunc = func(ctx context.Context, id string) error {
					return nil
				}
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "not found",
			id:   "nonexistent",
			setup: func(m *mockService) {
				m.deleteFunc = func(ctx context.Context, id string) error {
					return domain.ErrNotFound
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "internal error",
			id:   "url-id-1",
			setup: func(m *mockService) {
				m.deleteFunc = func(ctx context.Context, id string) error {
					return errors.New("db error")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockService{}
			tt.setup(mock)
			h := NewHandler(mock)
			r := NewRouter(h)

			req := httptest.NewRequest(http.MethodDelete, "/api/urls/"+tt.id, nil)
			ctx := chi.NewRouteContext()
			ctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
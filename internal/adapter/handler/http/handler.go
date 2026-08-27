package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/hiiamhip/nanoflux/internal/domain"
)

type URLService interface {
	CreateShortURL(ctx context.Context, original string) (string, error)
	ResolveURL(ctx context.Context, code string, ip string, userAgent string) (string, error)
	UpdateURL(ctx context.Context, id, newOriginal string) error
	DeleteURL(ctx context.Context, id string) error
}

type Handler struct {
	svc URLService
}

func NewHandler(svc URLService) *Handler {
	return &Handler{svc: svc}
}

type createRequest struct {
	URL string `json:"url"`
}

type createResponse struct {
	Code string `json:"code"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	code, err := h.svc.CreateShortURL(r.Context(), req.URL)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidURL):
			httpError(w, http.StatusBadRequest, "invalid url")
		default:
			httpError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, createResponse{Code: code})
}

func (h *Handler) Resolve(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	ip := r.RemoteAddr
	userAgent := r.UserAgent()

	original, err := h.svc.ResolveURL(r.Context(), code, ip, userAgent)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			httpError(w, http.StatusNotFound, "url not found")
		default:
			httpError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	http.Redirect(w, r, original, http.StatusFound)
}

type updateRequest struct {
	URL string `json:"url"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	err := h.svc.UpdateURL(r.Context(), id, req.URL)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			httpError(w, http.StatusNotFound, "url not found")
		case errors.Is(err, domain.ErrInvalidURL):
			httpError(w, http.StatusBadRequest, "invalid url")
		default:
			httpError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.svc.DeleteURL(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			httpError(w, http.StatusNotFound, "url not found")
		default:
			httpError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func httpError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

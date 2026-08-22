package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hiiamhip/nanoflux/internal/domain"
	"github.com/hiiamhip/nanoflux/internal/usecase"
)

type Handler struct {
	svc *usecase.URLService
}

func NewHandler(svc *usecase.URLService) *Handler {
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

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func httpError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

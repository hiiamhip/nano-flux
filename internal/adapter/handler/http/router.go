package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/api/urls", h.Create)
	r.Get("/{code}", h.Resolve)
	return r
}

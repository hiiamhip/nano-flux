package http

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	limiter := rate.NewLimiter(rate.Every(100*time.Millisecond), 20)
	r.Use(rateLimitMiddleware(limiter))

	r.Post("/api/urls", h.Create)
	r.Get("/{code}", h.Resolve)
	r.Put("/api/urls/{id}", h.Update)
	r.Delete("/api/urls/{id}", h.Delete)
	return r
}

func rateLimitMiddleware(limiter *rate.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	httpapi "github.com/hiiamhip/nanoflux/internal/adapter/handler/http"
	"github.com/hiiamhip/nanoflux/internal/adapter/repository/postgres"
	"github.com/hiiamhip/nanoflux/internal/adapter/shortener"
	"github.com/hiiamhip/nanoflux/internal/config"
	"github.com/hiiamhip/nanoflux/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	// adapter
	repo := postgres.NewPostgresRepo(pool)
	gen := shortener.NewBase62()

	// usecase
	svc := usecase.NewURLService(repo, gen)

	// handler
	h := httpapi.NewHandler(svc)

	// router
	r := httpapi.NewRouter(h)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	log.Printf("listening on :%s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server: %v", err)
	}
}

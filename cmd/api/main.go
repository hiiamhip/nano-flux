package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	httpapi "github.com/hiiamhip/nanoflux/internal/adapter/handler/http"
	"github.com/hiiamhip/nanoflux/internal/adapter/repository/postgres"
	"github.com/hiiamhip/nanoflux/internal/config"
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

	h := httpapi.NewHandler()
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

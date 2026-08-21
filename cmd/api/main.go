package main

import (
	"context"
	"log"

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

	_ = pool
}

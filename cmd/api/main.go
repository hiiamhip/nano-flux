package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// adapter
	repo := postgres.NewURLRepo(pool)
	gen := shortener.NewBase62()

	// usecase
	svc := usecase.NewURLService(repo, gen)

	// handler
	h := httpapi.NewHandler(svc)

	// router
	r := httpapi.NewRouter(h)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("listening on :%s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	pool.Close()
	log.Println("server exited gracefully")
}

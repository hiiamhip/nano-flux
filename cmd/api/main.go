package main

import (
	"log"

	"github.com/hiiamhip/nanoflux/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	_ = cfg
}

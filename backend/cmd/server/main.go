package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"podcast-summarizer/src/api"
	"podcast-summarizer/src/config"
	"podcast-summarizer/src/core/app/jobs"
	cacheinfra "podcast-summarizer/src/infra/cache"
	dbinfra "podcast-summarizer/src/infra/db"
	storageinfra "podcast-summarizer/src/infra/storage"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	pool, err := dbinfra.NewPostgresPool(ctx)
	if err != nil {
		log.Fatalf("postgres init failed: %v", err)
	}
	defer pool.Close()

	valkey := cacheinfra.NewValkeyClient()
	if err := cacheinfra.PingValkey(ctx, valkey); err != nil {
		log.Fatalf("valkey init failed: %v", err)
	}

	var storage jobs.ObjectUploader = &storageinfra.NoopUploader{}
	if cfg.R2Bucket != "" {
		if uploader, err := storageinfra.NewR2Uploader(ctx, cfg); err == nil {
			storage = uploader
		} else {
			log.Printf("warning: R2 uploader not initialized: %v", err)
		}
	}

	handler := api.NewRouter(api.Dependencies{
		Config:   cfg,
		Postgres: pool,
		Valkey:   valkey,
		Context:  ctx,
		Storage:  storage,
	})

	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("starting server on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}

	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
}

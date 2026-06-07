package main

import (
	"context"
	"log"
	"net/http"
	"strings"
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

	var deps api.Dependencies
	deps.Config = cfg
	deps.Context = ctx

	switch strings.ToLower(cfg.StorageDriver) {
	case "", "sqlite":
		sqliteDB, err := dbinfra.NewSQLite(ctx, cfg.SQLitePath)
		if err != nil {
			log.Fatalf("sqlite init failed: %v", err)
		}
		defer func() {
			_ = sqliteDB.Close()
		}()
		deps.SQLite = sqliteDB
		log.Printf("using sqlite storage at %s", cfg.SQLitePath)
	case "postgres":
		pool, err := dbinfra.NewPostgresPool(ctx)
		if err != nil {
			log.Fatalf("postgres init failed: %v", err)
		}
		defer pool.Close()
		deps.Postgres = pool
		if cfg.ValkeyURL == "" {
			log.Fatalf("VALKEY_URL is required when STORAGE_DRIVER=postgres")
		}
		valkey := cacheinfra.NewValkeyClient()
		if err := cacheinfra.PingValkey(ctx, valkey); err != nil {
			log.Fatalf("valkey init failed: %v", err)
		}
		deps.Valkey = valkey
	default:
		log.Fatalf("unsupported STORAGE_DRIVER %q", cfg.StorageDriver)
	}

	var storage jobs.ObjectUploader
	switch strings.ToLower(cfg.ObjectStorageDriver) {
	case "", "local":
		storage = storageinfra.NewLocalUploader(cfg.LocalStoragePath)
		log.Printf("using local object storage at %s", cfg.LocalStoragePath)
	case "r2":
		if uploader, err := storageinfra.NewR2Uploader(ctx, cfg); err == nil {
			storage = uploader
		} else {
			log.Fatalf("R2 uploader not initialized: %v", err)
		}
	case "none":
		storage = &storageinfra.NoopUploader{}
	default:
		log.Fatalf("unsupported OBJECT_STORAGE_DRIVER %q", cfg.ObjectStorageDriver)
	}

	deps.Storage = storage
	handler := api.NewRouter(deps)

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

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"podcast-summarizer/src/api"
	"podcast-summarizer/src/config"
	"podcast-summarizer/src/core/app/jobs"
	dbinfra "podcast-summarizer/src/infra/db"
	storageinfra "podcast-summarizer/src/infra/storage"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	var deps api.Dependencies
	deps.Config = cfg
	deps.Context = ctx

	sqliteDB, err := dbinfra.NewSQLite(ctx, cfg.SQLitePath)
	if err != nil {
		log.Fatalf("sqlite init failed: %v", err)
	}
	defer func() {
		_ = sqliteDB.Close()
	}()
	deps.SQLite = sqliteDB
	log.Printf("using sqlite storage at %s", cfg.SQLitePath)

	var storage jobs.ObjectUploader = storageinfra.NewLocalUploader(cfg.LocalStoragePath)
	log.Printf("using local object storage at %s", cfg.LocalStoragePath)

	deps.Storage = storage
	handler := api.NewRouter(deps)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	server := &http.Server{
		Addr:         port,
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

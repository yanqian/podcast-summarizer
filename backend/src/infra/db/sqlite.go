package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// NewSQLite opens the local demo database and applies the embedded schema.
func NewSQLite(ctx context.Context, path string) (*sql.DB, error) {
	if path == "" {
		path = "data/podcast.db"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create sqlite dir: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	if err := applySQLiteSchema(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func applySQLiteSchema(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		`PRAGMA foreign_keys = ON;`,
		`PRAGMA journal_mode = WAL;`,
		`PRAGMA busy_timeout = 5000;`,
		`CREATE TABLE IF NOT EXISTS podcast_source (
			id TEXT PRIMARY KEY,
			url TEXT NOT NULL UNIQUE,
			title TEXT,
			description TEXT,
			duration_seconds INTEGER,
			audio_url TEXT,
			transcript_url TEXT,
			has_transcript INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS transcript_paragraph (
			id TEXT PRIMARY KEY,
			podcast_id TEXT NOT NULL REFERENCES podcast_source(id) ON DELETE CASCADE,
			order_index INTEGER NOT NULL,
			text TEXT NOT NULL,
			timestamp_seconds INTEGER,
			source TEXT NOT NULL DEFAULT 'generated',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (podcast_id, order_index)
		);`,
		`CREATE TABLE IF NOT EXISTS summary_paragraph (
			id TEXT PRIMARY KEY,
			transcript_paragraph_id TEXT NOT NULL REFERENCES transcript_paragraph(id) ON DELETE CASCADE,
			summary_text TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (transcript_paragraph_id)
		);`,
		`CREATE TABLE IF NOT EXISTS processing_job (
			id TEXT PRIMARY KEY,
			podcast_id TEXT NOT NULL REFERENCES podcast_source(id) ON DELETE CASCADE,
			type TEXT NOT NULL,
			status TEXT NOT NULL,
			started_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			completed_at TEXT,
			duration_ms INTEGER,
			error_message TEXT,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS processing_job_podcast_created_idx
			ON processing_job (podcast_id, created_at DESC);`,
		`CREATE TABLE IF NOT EXISTS job_lock (
			key TEXT PRIMARY KEY,
			expires_at TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
	}
	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("apply sqlite schema: %w", err)
		}
	}
	return nil
}

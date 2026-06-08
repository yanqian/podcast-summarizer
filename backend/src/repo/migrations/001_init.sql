-- SQLite schema for the local-first podcast summarizer runtime.

PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;
PRAGMA busy_timeout = 5000;

CREATE TABLE IF NOT EXISTS podcast_source (
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
);

CREATE TABLE IF NOT EXISTS transcript_paragraph (
    id TEXT PRIMARY KEY,
    podcast_id TEXT NOT NULL REFERENCES podcast_source(id) ON DELETE CASCADE,
    order_index INTEGER NOT NULL,
    text TEXT NOT NULL,
    timestamp_seconds INTEGER,
    source TEXT NOT NULL DEFAULT 'generated',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (podcast_id, order_index)
);

CREATE TABLE IF NOT EXISTS summary_paragraph (
    id TEXT PRIMARY KEY,
    transcript_paragraph_id TEXT NOT NULL REFERENCES transcript_paragraph(id) ON DELETE CASCADE,
    summary_text TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (transcript_paragraph_id)
);

CREATE TABLE IF NOT EXISTS processing_job (
    id TEXT PRIMARY KEY,
    podcast_id TEXT NOT NULL REFERENCES podcast_source(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    started_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TEXT,
    duration_ms INTEGER,
    error_message TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS processing_job_podcast_created_idx
    ON processing_job (podcast_id, created_at DESC);

CREATE TABLE IF NOT EXISTS job_lock (
    key TEXT PRIMARY KEY,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

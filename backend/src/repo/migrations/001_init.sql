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

CREATE TABLE IF NOT EXISTS episode (
    id TEXT PRIMARY KEY,
    podcast_url TEXT NOT NULL UNIQUE,
    title TEXT,
    description TEXT,
    duration_seconds INTEGER,
    audio_url TEXT,
    transcript_url TEXT,
    has_transcript INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO episode (
    id,
    podcast_url,
    title,
    description,
    duration_seconds,
    audio_url,
    transcript_url,
    has_transcript,
    created_at,
    updated_at
)
SELECT
    id,
    url,
    title,
    description,
    duration_seconds,
    audio_url,
    transcript_url,
    has_transcript,
    created_at,
    updated_at
FROM podcast_source
;

CREATE TABLE IF NOT EXISTS transcript_paragraph (
    id TEXT PRIMARY KEY,
    podcast_id TEXT NOT NULL REFERENCES episode(id) ON DELETE CASCADE,
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
    podcast_id TEXT NOT NULL REFERENCES episode(id) ON DELETE CASCADE,
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

CREATE TABLE IF NOT EXISTS audio_chunk (
    id TEXT PRIMARY KEY,
    episode_id TEXT NOT NULL REFERENCES episode(id) ON DELETE CASCADE,
    processing_job_id TEXT REFERENCES processing_job(id) ON DELETE SET NULL,
    order_index INTEGER NOT NULL,
    file_path TEXT NOT NULL,
    start_seconds REAL,
    end_seconds REAL,
    duration_seconds REAL,
    byte_size INTEGER,
    checksum TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (episode_id, order_index)
);

CREATE INDEX IF NOT EXISTS audio_chunk_episode_order_idx
    ON audio_chunk (episode_id, order_index);

CREATE TABLE IF NOT EXISTS transcript_segment (
    id TEXT PRIMARY KEY,
    episode_id TEXT NOT NULL REFERENCES episode(id) ON DELETE CASCADE,
    audio_chunk_id TEXT REFERENCES audio_chunk(id) ON DELETE SET NULL,
    order_index INTEGER NOT NULL,
    start_seconds REAL,
    end_seconds REAL,
    text TEXT NOT NULL,
    provider TEXT,
    model TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (episode_id, order_index)
);

CREATE INDEX IF NOT EXISTS transcript_segment_episode_order_idx
    ON transcript_segment (episode_id, order_index);

CREATE TABLE IF NOT EXISTS summary_segment (
    id TEXT PRIMARY KEY,
    episode_id TEXT NOT NULL REFERENCES episode(id) ON DELETE CASCADE,
    order_index INTEGER NOT NULL,
    text TEXT NOT NULL,
    provider TEXT,
    model TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (episode_id, order_index)
);

CREATE INDEX IF NOT EXISTS summary_segment_episode_order_idx
    ON summary_segment (episode_id, order_index);

CREATE TABLE IF NOT EXISTS transcript_summary_mapping (
    id TEXT PRIMARY KEY,
    episode_id TEXT NOT NULL REFERENCES episode(id) ON DELETE CASCADE,
    summary_segment_id TEXT NOT NULL REFERENCES summary_segment(id) ON DELETE CASCADE,
    transcript_segment_id TEXT NOT NULL REFERENCES transcript_segment(id) ON DELETE CASCADE,
    source_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (summary_segment_id, transcript_segment_id)
);

CREATE INDEX IF NOT EXISTS transcript_summary_mapping_episode_idx
    ON transcript_summary_mapping (episode_id, summary_segment_id, source_order);

CREATE TABLE IF NOT EXISTS job_lock (
    key TEXT PRIMARY KEY,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

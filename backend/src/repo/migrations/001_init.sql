-- schema for podcast summarizer

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS podcast_source (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url TEXT NOT NULL UNIQUE,
    title TEXT,
    description TEXT,
    duration_seconds INTEGER,
    has_transcript BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS transcript_paragraph (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    podcast_id UUID NOT NULL REFERENCES podcast_source(id) ON DELETE CASCADE,
    order_index INTEGER NOT NULL,
    text TEXT NOT NULL,
    timestamp_seconds INTEGER,
    source TEXT NOT NULL DEFAULT 'provided',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (podcast_id, order_index)
);

CREATE TABLE IF NOT EXISTS summary_paragraph (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transcript_paragraph_id UUID NOT NULL REFERENCES transcript_paragraph(id) ON DELETE CASCADE,
    summary_text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (transcript_paragraph_id)
);

CREATE TABLE IF NOT EXISTS processing_job (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    podcast_id UUID NOT NULL REFERENCES podcast_source(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('ingest', 'transcribe', 'summarize')),
    status TEXT NOT NULL CHECK (status IN ('queued', 'running', 'succeeded', 'failed')),
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,
    duration_ms INTEGER,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Ensure only one active job (queued or running) per podcast/type at a time
CREATE UNIQUE INDEX IF NOT EXISTS processing_job_active_idx
    ON processing_job (podcast_id, type)
    WHERE status IN ('queued', 'running');

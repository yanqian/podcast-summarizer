# Backend Skeleton

Structure initialized per plan with clean architecture boundaries:
- `src/api`: HTTP handlers, DTOs, routing.
- `src/core`: use cases/services (transcription, summarization, ingest, jobs).
- `src/adapters`: external integrations for transcript fetch, transcription, summarization.
- `src/repo`: Postgres implementations.
- `src/config`: configuration loading.
- `tests/`: unit, integration, and contract suites.

Prerequisites:
- Go 1.22+
- ffmpeg available on PATH (used for audio chunking). Set `FFMPEG_PATH` if not simply `ffmpeg`.

Next steps: add router, middleware, repositories, and adapters per tasks.md.

Current behavior:
- Ingest extracts track ID, fetches metadata via iTunes lookup, enqueues job.
- Job manager streams transcript chunks over SSE; saves transcript/summaries to Postgres (or in-memory fallback).
- Transcript fetcher/download + ffmpeg chunker + transcriber/summarizer adapters are wired; HTTP adapters used when `TRANSCRIBE_URL/KEY` and `SUMMARIZE_URL/KEY` are set, otherwise stubs.

Tests:
- `GOCACHE=$(pwd)/.gocache go test ./...` (contract/integration with in-memory fallbacks).

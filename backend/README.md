# Backend

Structure initialized per plan with clean architecture boundaries:
- `src/api`: HTTP handlers, DTOs, routing.
- `src/core`: use cases/services (transcription, summarization, ingest, jobs).
- `src/adapters`: external integrations for transcript fetch, transcription, summarization.
- `src/repo`: SQLite repository implementations.
- `src/config`: configuration loading.
- `tests/`: unit, integration, and contract suites.

Prerequisites:
- Go 1.25+
- ffmpeg available on PATH (used for audio chunking). Set `FFMPEG_PATH` if not simply `ffmpeg`.

Local runtime:
- The server uses SQLite only.
- The database file is created automatically at `data/podcast.db` unless `SQLITE_PATH` is set.
- Generated audio/chunk artifacts are copied to local file storage at `data/storage` unless `LOCAL_STORAGE_PATH` is set.

Current behavior:
- Ingest extracts track ID, fetches metadata via iTunes lookup, enqueues job.
- Job manager streams transcript chunks over SSE; saves transcript/summaries to SQLite.
- Transcript fetcher/download + ffmpeg chunker + transcriber/summarizer adapters are wired; HTTP adapters used when `TRANSCRIBE_URL/KEY` and `SUMMARIZE_URL/KEY` are set, otherwise stubs.

Tests:
- `GOCACHE=$(pwd)/.gocache go test ./...`

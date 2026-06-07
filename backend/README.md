# Backend

Structure initialized per plan with clean architecture boundaries:
- `src/api`: HTTP handlers, DTOs, routing.
- `src/core`: use cases/services (transcription, summarization, ingest, jobs).
- `src/adapters`: external integrations for transcript fetch, transcription, summarization.
- `src/repo`: SQLite and Postgres repository implementations.
- `src/config`: configuration loading.
- `tests/`: unit, integration, and contract suites.

Prerequisites:
- Go 1.25+
- ffmpeg available on PATH (used for audio chunking). Set `FFMPEG_PATH` if not simply `ffmpeg`.

Default local/demo mode:
- The server uses SQLite by default (`STORAGE_DRIVER=sqlite`).
- The database file is created automatically at `data/podcast.db` unless `SQLITE_PATH` is set.
- Generated audio/chunk artifacts are copied to local file storage at `data/storage` unless `LOCAL_STORAGE_PATH` is set.
- Postgres and Valkey are not required for local demo use.

Current behavior:
- Ingest extracts track ID, fetches metadata via iTunes lookup, enqueues job.
- Job manager streams transcript chunks over SSE; saves transcript/summaries to SQLite by default.
- Transcript fetcher/download + ffmpeg chunker + transcriber/summarizer adapters are wired; HTTP adapters used when `TRANSCRIBE_URL/KEY` and `SUMMARIZE_URL/KEY` are set, otherwise stubs.

Optional cloud mode:
- Set `STORAGE_DRIVER=postgres` and `POSTGRES_URL` to use Postgres.
- Set `VALKEY_URL` to use Valkey for distributed job locking. Postgres mode requires Valkey; SQLite mode uses SQLite-backed job locking.
- Set `OBJECT_STORAGE_DRIVER=r2` plus R2 credentials to store generated artifacts in Cloudflare R2.

Tests:
- `GOCACHE=$(pwd)/.gocache go test ./...`

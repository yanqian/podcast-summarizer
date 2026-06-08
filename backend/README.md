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
- Downloaded episode audio is copied to `data/storage/<episode-id>/original.mp3` unless `LOCAL_STORAGE_PATH` is set.
- ffmpeg chunks are copied to `data/storage/<episode-id>/chunks/chunk-001.mp3`, `chunk-002.mp3`, and so on, with matching SQLite `audio_chunk` rows in stable order.
- SQLite stores episode metadata, processing jobs, audio chunk records, transcript segments, summary segments, and transcript-to-summary mappings. The database and generated files live under `data/` by default and are ignored by git.

Current behavior:
- Ingest extracts track ID, fetches metadata via iTunes lookup, enqueues job.
- Job manager streams transcript chunks over SSE; saves transcript/summaries to SQLite.
- Transcript fetcher/download + ffmpeg chunker + transcriber/summarizer adapters are wired; HTTP adapters used when `TRANSCRIBE_URL/KEY` and `SUMMARIZE_URL/KEY` are set, otherwise stubs.

Tests:
- `GOCACHE=$(pwd)/.gocache go test ./...`

# Quickstart: Podcast Summary UI

Date: 2025-11-29  
Branch: 001-podcast-summary-ui

## Prerequisites

- Go 1.25+, Node.js + npm
- ffmpeg on PATH for audio transcription flows

## Setup

1. Install frontend deps: `cd frontend && npm ci`
2. Install backend deps: `cd backend && go mod download`
3. Optional: copy `backend/.env.example` to `backend/.env` and fill external adapter keys.

## Run

1. Backend: `cd backend && go run ./cmd/server` (exposes REST API per contracts/openapi.yaml)
   - Uses SQLite at `backend/data/podcast.db` unless `SQLITE_PATH` is set.
   - Defaults to local file storage at `backend/data/storage` for generated audio/chunk artifacts.
   - Uses SSE at `/api/streams/transcript/{jobId}` for chunk streaming.
2. Frontend: `cd frontend && npm run dev` (Vite dev server)
3. Open app, paste a podcast URL, observe streaming transcript chunks, and view aligned transcript/summary when complete.

## Tests

- Backend: `GOCACHE=$(pwd)/.gocache go test ./...`
- Frontend unit: `npm test` (uses vitest/jsdom; excludes e2e)
- Frontend e2e: Playwright removed; add back later if e2e coverage needed.
- Backend: `go test ./...` (unit, contract, and lightweight integration suites)
- Contract tests: ensure responses align with `contracts/openapi.yaml`

## Notes

- Performance budgets: UI interactions p95 <200ms; existing transcripts end-to-end <10s; transcription +
  summary for ≤60m episodes <5m for 90% cases.
- Logs: instrument transcript fetch/transcribe/summarize durations; surface job durations in status API.
- Placeholder perf checks: record ingest job durations in logs, capture p95 for streaming chunk arrival when
  running against real transcription/summarization services; verify export latency under 1s for small transcripts.
- Security: ingest validates URL scheme/length and caps request body; errors returned to clients are generic
  to avoid leaking internals.

## Current Status

- SSE streaming implemented with stubbed transcription/summarization; replace stubs with real services before release.
- SQLite-only demo storage implemented.
- Frontend deps installed via npm; backend tests pass; frontend unit test runs via vitest; no e2e runner.

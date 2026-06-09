# Quickstart: Podcast Summary UI

Date: 2025-11-29  
Branch: 001-podcast-summary-ui

## Prerequisites

- Go 1.25+, Node.js + npm
- ffmpeg on PATH for audio transcription flows

## Setup

1. Install frontend deps: `cd frontend && npm ci`
2. Install backend deps: `cd backend && go mod download`
3. Optional: copy `backend/.env.example` to `backend/.env` and set OpenAI keys for live transcription and summarization.

## Run

1. Backend: `cd backend && go run ./cmd/server`
   - Uses SQLite at `backend/data/podcast.db` unless `SQLITE_PATH` is set.
   - Defaults to local file storage at `backend/data/storage` for generated audio/chunk artifacts.
   - Uses deterministic local stubs when `OPENAI_API_KEY` is unset.
2. Frontend: `cd frontend && npm run dev` (Vite dev server)
3. Open the app. The Demo screen lists local episodes and renders segment-to-summary mappings; switch to Admin to submit podcast URLs and refresh processing status.

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
- Placeholder perf checks: record ingest job durations in logs, capture p95 for Admin status refresh and Demo detail rendering when
  running against real transcription/summarization services; verify export endpoint latency under 1s for small transcripts.
- Security: ingest validates URL scheme/length and caps request body; errors returned to clients are generic
  to avoid leaking internals.

## Current Status

- Demo/Admin frontend flow implemented with stubbed no-key transcription/summarization and OpenAI-backed live mode when configured.
- SQLite-only demo storage implemented.
- Frontend deps installed via npm; backend tests pass; frontend unit test runs via vitest; no e2e runner.

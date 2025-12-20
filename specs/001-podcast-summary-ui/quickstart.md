# Quickstart: Podcast Summary UI

Date: 2025-11-29  
Branch: 001-podcast-summary-ui

## Prerequisites

- Go (latest stable), Node.js + pnpm or npm, PostgreSQL, Valkey
- Environment variables for DB/cache/backends configured (URLs/credentials)

## Setup

1. Install frontend deps: `cd frontend && pnpm install`
2. Install backend deps: `cd backend && go mod download`
3. Start Postgres and Valkey locally (containers or services)
4. Apply initial migrations/seeds for podcast tables

## Run

1. Backend: `cd backend && go run ./cmd/server` (exposes REST API per contracts/openapi.yaml)
   - Requires Postgres/Valkey reachable via env vars; uses SSE at `/api/streams/transcript/{jobId}` for chunk streaming.
2. Frontend: `cd frontend && npm run dev` (Vite dev server)
3. Open app, paste a podcast URL, observe streaming transcript chunks, and view aligned transcript/summary when complete.

## Tests

- Backend: `GOCACHE=$(pwd)/.gocache go test ./...`
- Frontend unit: `npm test` (uses vitest/jsdom; excludes e2e)
- Frontend e2e: Playwright removed; add back later if e2e coverage needed.
- Backend: `go test ./...` (unit) and integration suite pointing to Postgres/Valkey test instances
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
- Frontend deps installed via npm; backend tests pass; frontend unit test runs via vitest; no e2e runner.

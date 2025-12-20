# Frontend Skeleton

Initialized directories for Vite + React + TypeScript + Tailwind + shadcn/ui:
- `src/components`, `src/pages`, `src/hooks`, `src/services`, `src/styles`
- `tests/unit`

Current behavior:
- Accepts podcast URL, calls ingest, subscribes to SSE stream at `/api/streams/transcript/{jobId}`, renders streaming transcript chunks, then fetches aligned transcript/summaries when stream completes.
- Status banner covers idle/loading/streaming/error/success states.

Tests:
- `npm test` runs vitest unit tests (UrlInput). E2E removed; reintroduce Playwright if needed.

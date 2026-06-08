# Frontend

Initialized directories for Vite + React + TypeScript:
- `src/components`, `src/pages`, `src/hooks`, `src/services`, `src/styles`
- `tests/unit`

Current behavior:
- Accepts podcast URL, calls ingest, subscribes to SSE stream at `/api/streams/transcript/{jobId}`, renders streaming transcript chunks, then fetches aligned transcript/summaries when stream completes.
- Status banner covers idle/loading/streaming/error/success states.

Configuration:
- `VITE_API_BASE_URL` defaults to `http://localhost:8080`.
- For local demos, set `VITE_API_BASE_URL` only when the backend is not listening on `http://localhost:8080`.
- The Docker image bakes `VITE_API_BASE_URL=http://localhost:8080` into the static bundle and serves the app on container port `8080`.
- The frontend container health check uses `/health`; it does not contact external services.

Tests:
- `npm test` runs Vitest unit tests for the submission flow and transcript-summary viewer.

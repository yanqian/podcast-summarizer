# Frontend

Initialized directories for Vite + React + TypeScript:
- `src/components`, `src/pages`, `src/hooks`, `src/services`, `src/styles`
- `tests/unit`

Current behavior:
- Opens on the Demo screen, which lists local episodes and renders transcript segments beside mapped summary segments from the current episode detail API.
- Provides an Admin screen, reachable through in-app navigation, for podcast URL submission, duplicate episode detection, processing status, errors, and manual refresh.
- Uses the backend HTTP episode list, detail, ingest, and status endpoints. The active frontend contract does not use SSE streaming, `/view`, export controls, or paragraph compatibility data.

Configuration:
- `VITE_API_BASE_URL` defaults to `http://localhost:8080`.
- For local demos, set `VITE_API_BASE_URL` only when the backend is not listening on `http://localhost:8080`.
- The Docker image bakes `VITE_API_BASE_URL=http://localhost:8080` into the static bundle and serves the app on container port `8080`.
- The frontend container health check uses `/health`; it does not contact external services.

Tests:
- `npm test` runs Vitest unit tests for the two-screen shell, submission flow, and segment-based transcript-summary viewer.
